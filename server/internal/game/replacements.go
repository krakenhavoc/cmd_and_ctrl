package game

import (
	"errors"

	"github.com/google/uuid"
)

// replacements.go is the S17 CR 614/616 replacement-effects engine.
//
// Replacement effects watch for candidate events (draw, zone move,
// counter placement, life change, damage, step transition) and
// substitute a different event — or cancel it — before it happens.
// Pre-event interception, unlike S14 listeners which are post-event.
//
// The engine is declarative: cards register a ReplacementEffect on
// effects.Spec.Replacements, the engine walks applicable effects at
// mutation time, picks one to apply (or queues a CR 616 prompt when
// multiple apply), applies it, and iterates (CR 616.1 iterative
// apply-loop). CR 614.5 once-per-event tracking prevents the same
// effect from firing twice on one event.
//
// Pipeline entry points (in mutations.go + game.go):
//   - drawCardLocked → RepEventDraw
//   - MoveCardByIDAsCommander → RepEventMove (carries EntersTapped,
//     EntersWithCounters, asCommanderMove breadcrumb)
//   - ChangePlayerLife → RepEventLife
//   - MarkDamage / MarkCombatDamage → RepEventDamage
//   - AddCounter → RepEventCounter
//   - runStepEntryHooksLocked → RepEventStepTransition
//
// The six *ForEffect helpers in effect_api.go route through the same
// pipeline so catalog-driven mutations fire replacements too.
//
// Added in S17 sub-PR 2 — zero catalog replacements registered; only
// the commander-zone built-in (refactored from S13.1) and test-only
// injections. Subsequent sub-PRs populate catalog cards that exercise
// the engine.

// ReplacementEventKind narrows the meaningful fields on a
// ReplacementEvent. Values:
//
//	"draw"    — RepEventDraw    — DrawPlayer
//	"move"    — RepEventMove    — CardID, OldZone, NewZone, NewZoneOwner, EntersTapped, EntersWithCounters, asCommanderMove
//	"counter" — RepEventCounter — CounterTarget, CounterName, CounterDelta
//	"life"    — RepEventLife    — LifePlayer, LifeDelta
//	"damage"  — RepEventDamage  — DamageSource, DamageTarget, DamageAmount, IsCombatDamage
//	"step"    — RepEventStepTransition — StepTransitionStep, StepTransitionSeat
type ReplacementEventKind string

const (
	RepEventDraw           ReplacementEventKind = "draw"
	RepEventMove           ReplacementEventKind = "move"
	RepEventCounter        ReplacementEventKind = "counter"
	RepEventLife           ReplacementEventKind = "life"
	RepEventDamage         ReplacementEventKind = "damage"
	RepEventStepTransition ReplacementEventKind = "step"
)

// ReplacementEventID is the per-event key used by the once-per-event
// tracking map (CR 614.5). Minted by applyReplacementsLocked on first
// entry; stable across the CR 616.1 iterative apply-loop for a single
// event so the tracking map stays consistent.
type ReplacementEventID uint64

// ReplacementEffectID is the per-instance ID for one active
// replacement effect. Stable across the apply-loop for a single
// event. For built-ins it's minted once at NewGame; for catalog
// replacements it's minted on gather keyed by (source card
// InstanceID, index into the card's Replacements slice).
type ReplacementEffectID uint64

// builtinReplacementIDBase is the offset subtracted from a built-in's
// engine-assigned ID. Built-ins' IDs live in the top uint64 half so
// they never collide with catalog-assigned IDs.
const builtinReplacementIDBase ReplacementEffectID = 1 << 62

// ReplacementEvent is the mutable pre-event value passed through
// the replacement pipeline. Tagged-union shape (discriminated by
// Kind) mirrors the Event struct used by EmitEvent, but with
// Canceled + per-kind mutable fields that replacements can rewrite
// before the underlying mutation runs.
type ReplacementEvent struct {
	// ID is minted by applyReplacementsLocked on first entry.
	// Stable across the CR 616.1 iterative apply-loop. Keys the
	// once-per-event tracking map.
	ID ReplacementEventID

	// Kind narrows the meaningful fields below. Callers read and
	// mutate only the fields valid for the kind.
	Kind ReplacementEventKind

	// Canceled flips true when a replacement suppresses the event
	// entirely (CR 614.10 — "instead" with a null replacement).
	// Pipeline functions observe this and skip the underlying
	// mutation + the post-event EmitEvent.
	Canceled bool

	// Actor is the player responsible for the event (drawer for
	// Draw, mover-invoker for Move, etc.). Carried verbatim into
	// the emitted Event when the event fires.
	Actor uuid.UUID

	// Source is the card that caused the event — the spell /
	// ability / permanent that initiated it. uuid.Nil when the
	// event has no specific source (admin-driven mutations).
	Source uuid.UUID

	// --- RepEventDraw fields ---

	// DrawPlayer is the player drawing. Replacements can rewrite
	// (redirect, "if you would draw, opponent mills instead" style)
	// or Cancel (skip the draw entirely).
	DrawPlayer uuid.UUID

	// --- RepEventMove fields ---

	// CardID is the card being moved.
	CardID uuid.UUID

	// OldZone / NewZone / NewZoneOwner describe the motion.
	// Replacements can rewrite NewZone (the CR 903.9 commander-zone
	// built-in, Stone of Erech's graveyard → exile, etc.). No discard
	// builds a RepEventMove yet: every discard moves the card directly
	// and bypasses this pipeline (#650).
	OldZone      ZoneKind
	NewZone      ZoneKind
	NewZoneOwner uuid.UUID

	// EntersTapped is mutated by enters-tapped replacements
	// (Kismet). Only meaningful when NewZone == ZoneBattlefield.
	// The battlefield-entry path reads this and sets Card.Tapped
	// before emitting EventETB.
	EntersTapped bool

	// EntersAsCopyOf is the CR 707 copy a permanent enters wearing —
	// the copiable values settled by a CopySelector replacement,
	// with the card's "except" clause already applied. nil for the
	// ~everything that enters as itself. Only meaningful when
	// NewZone == ZoneBattlefield; the entry path materialises it
	// onto the card BEFORE EventETB fires, so no trigger ever sees
	// the permanent as its own printed self. Added in S16.5 (#159).
	EntersAsCopyOf *PrintedValues

	// copySourceID names the permanent EntersAsCopyOf was taken
	// from. Unexported because the catalog has no business reading
	// it: it exists so the entry path can pick up the source's
	// card-carried ability slices (the token case, which
	// PrintedValues cannot carry — see copy.go) and so the copy
	// event names what was copied.
	copySourceID uuid.UUID

	// stackItem is the resolving spell's StackItem, carried across a
	// paused entry so the resume can finish the two jobs only stack
	// resolution does: attaching a resolved Aura to what it targeted
	// (CR 303.4a) and queueing evoke's sacrifice trigger (CR
	// 702.74a). Unexported engine plumbing.
	//
	// Set by resolveTopOfStackLocked, which is what makes that entry
	// site entryResumable. Before it existed, a permanent spell
	// whose entry queued any prompt was never pushed at all.
	stackItem *StackItem

	// EntersWithCounters is the map of counter name → count applied
	// BEFORE EventETB fires (Hangarback Walker "enters with X +1/+1
	// counters", etc.). Applied by the battlefield-entry path under
	// the same lock before the ETB event.
	EntersWithCounters map[string]int

	// entryResumable is an unexported breadcrumb meaning "if this
	// entry pauses for a prompt, the generic resume path may finish
	// it" (executeEntryToBattlefieldLocked). Set by the land-play
	// branch of CastSpell and — since S16.5, so Clone can ask what
	// to copy — by stack resolution, which hands the resume its
	// StackItem so the Aura attach and evoke's sacrifice trigger
	// survive the pause.
	//
	// Off by default on purpose. The remaining battlefield-entry
	// sites do things the generic push can't: an exile→battlefield
	// return mints a new InstanceID (CR 400.7), and the library-
	// search path owes its caller a shuffle. Finishing those
	// generically would silently drop the new object identity or
	// leak library order, which is worse than leaving them exactly
	// as they are — they still never pause today.
	//
	// An effect that WOULD pause consults this before prompting: a
	// pay-life entry choice on an unflagged event takes the un-paid
	// branch rather than stranding the card in its old zone (see
	// offerEntryLifePaymentLocked). Any entry site that grows a
	// faithful resume should set this and inherit the prompt.
	entryResumable bool

	// zoneRoute is the exit half's answer to entryResumable: the
	// per-destination bookkeeping (to the bottom of the library, face
	// down in exile, this was a mill, this was a counterspell) that a
	// paused move has to carry across the pause so the resume can
	// finish it exactly as the mover asked. Set by
	// routeCardToZoneLocked and read by executeZoneRouteLocked; a
	// non-nil value is what makes a RepEventMove resumable on the
	// EXIT side, the way entryResumable does on the entry side.
	//
	// Unexported engine plumbing — the catalog never sets or reads
	// it. Added in #529 so CR 903.9 could move down to the primitive.
	zoneRoute *zoneRoute

	// asCommanderMove is an unexported breadcrumb set by
	// MoveCardByIDAsCommander when the caller flagged the move as
	// a commander-initiated one. It is NOT a gate on the CR 903.9
	// built-in — that gate was dropped in #171 and the replacement
	// has been destination-only ever since. It survives as a routing
	// flavor flag on the manual move_card action. Unexported because
	// the catalog should never read or set it.
	asCommanderMove bool

	// --- RepEventCounter fields ---

	// CounterTarget / CounterName / CounterDelta are the counter
	// operation. Doubling Season doubles Delta; Hardened Scales adds
	// 1 when Name is "+1/+1"; Solemnity would set Canceled on +1/+1.
	CounterTarget uuid.UUID
	CounterName   string
	CounterDelta  int

	// --- RepEventLife fields ---

	// LifePlayer / LifeDelta describe the life change.
	LifePlayer uuid.UUID
	LifeDelta  int

	// lifeTail is the life half's answer to damageTail: the rest of
	// the effect that asked for this change, run with the amount that
	// actually moved once the pipeline settles. Set by the *Then*
	// entry points in effect_api.go and read only by
	// runLifeTailLocked, which both the inline path and the CR 616
	// resume reach.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	// Added in #793, where "each opponent loses X life, you gain life
	// equal to the life lost this way" read the life total back on the
	// next line and saw nothing whenever the loss paused on a CR 616
	// prompt. See life_tail.go.
	lifeTail *lifeTail

	// --- RepEventDamage fields ---

	// DamageSource / DamageTarget / DamageAmount describe the
	// damage event. Target is a card ID (creature / planeswalker)
	// or player ID; kind differentiates. IsCombatDamage flags
	// damage from the combat-damage step so Fog-class effects can
	// key on it without catching non-combat damage.
	DamageSource   uuid.UUID
	DamageTarget   uuid.UUID
	DamageAmount   int
	IsCombatDamage bool

	// damageTail is the damage half's answer to zoneRoute: what the
	// entry point still owes once the pipeline settles the amount —
	// whether the target is a player or a permanent, and the snapshot
	// of the source's deathtouch / lifelink / commander state that the
	// riders need. Set by every damage entry point and read only by
	// applyResolvedDamageLocked, which both the inline path and the
	// CR 616 resume go through.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	// Added in #694, where the resume's hand-rolled copy of the tail
	// was dropping player life loss, the CR 120.3 split, lifelink,
	// deathtouch and commander damage. See damage_tail.go.
	damageTail *damageTail

	// --- RepEventStepTransition fields ---

	// StepTransitionStep is the step being entered (StepUntap,
	// StepUpkeep, etc.). Stasis watches for StepUntap.
	StepTransitionStep Step
	// StepTransitionSeat is the seat whose step is being entered.
	StepTransitionSeat int

	// mustSettleNow marks an event that CANNOT pause: the pipeline
	// must reach a settled answer before applyReplacementsLocked
	// returns, because the caller has no resume and no way to be
	// rewound once it has.
	//
	// One thing sets it today: paying life as a cost (CR 118.3).
	// CR 601.2h pays a spell's costs as one indivisible step of
	// casting it and CR 601.2 rewinds the announcement if they cannot
	// all be paid, so a CR 616 ordering prompt in the middle leaves a
	// spell on the stack with its cost half paid. See
	// payLifeAsCostLocked in life_tail.go for the full argument.
	//
	// What it costs the affected player is the CR 616 ordering choice
	// and any CR 614.10 "may" on the event: the apply-loop applies the
	// gathered order inline (the escape an eliminated chooser has
	// always taken) and skips anything that would ask a question
	// un-applied. Weaker than printed on the "may", arbitrary but
	// deterministic on the ordering, and never a wedged table.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	mustSettleNow bool
}

// Cancel marks the event suppressed. Pipeline functions observing
// ev.Canceled skip the underlying mutation and the post-event
// EmitEvent.
func (ev *ReplacementEvent) Cancel() { ev.Canceled = true }

// AddCounterAtETB appends a counter-entry directive applied before
// EventETB fires. Only meaningful on RepEventMove with
// NewZone == ZoneBattlefield. Idempotent per name — adds to the
// existing count. Used by Hangarback-Walker-style "enters with X
// +1/+1 counters" effects.
func (ev *ReplacementEvent) AddCounterAtETB(name string, n int) {
	if name == "" || n == 0 {
		return
	}
	if ev.EntersWithCounters == nil {
		ev.EntersWithCounters = make(map[string]int)
	}
	ev.EntersWithCounters[name] += n
}

// ReplacementEffect declares one continuous replacement-effect
// contribution from a card or the engine. Structural sibling of
// StaticAbility (S16).
//
// Cards populate effects.Spec.Replacements with these. The engine
// gathers them at apply time, filters by Watches + AppliesTo, and
// fires Replace. CR 616.1 iteration + CR 614.5 once-per-event
// tracking + CR 616 order-choose prompt are all enforced centrally;
// cards stay declarative.
type ReplacementEffect struct {
	// Watches is the set of event kinds this effect is interested
	// in. Cheap pre-filter — if ev.Kind isn't in Watches, AppliesTo
	// is skipped. Empty Watches matches any kind (rarely useful).
	//
	// Uses the existing EventKind enum from events.go rather than
	// ReplacementEventKind so built-in EventKind values (e.g.
	// EventCounterPlaced) double as both post-event listener keys
	// and pre-event replacement keys.
	Watches []EventKind

	// AppliesTo is the predicate — does this candidate event match
	// what the effect replaces? Source is the card hosting the
	// effect (nil for built-in replacements like commander zone).
	// Returning false short-circuits without calling Replace.
	AppliesTo func(ev *ReplacementEvent, g *Game, source *Card) bool

	// Replace runs when the effect fires. Mutates *ev in place to
	// substitute a different event, or calls ev.Cancel() to
	// suppress the event entirely. Return error for diagnostic /
	// effect-error logging; a non-nil error does NOT cancel the
	// event (treat as "replacement misbehaved, let the original
	// through" — symmetric with fireEffectResolverLocked).
	Replace func(ev *ReplacementEvent, g *Game, source *Card) error

	// Controller returns the player who controls the effect. Drives
	// CR 616 affected-player ordering (the affected player picks
	// order among replacements they control + replacements the
	// event's affected player controls). For built-in replacements
	// it's typically the event's affected player.
	Controller func(ev *ReplacementEvent, g *Game, source *Card) uuid.UUID

	// SelfReplacement flags effects that replace an event affecting
	// their own source (CR 614.5). S17 treats all fired effects
	// identically via the once-per-event map; this field remains
	// on the struct for documentation and future hooks (e.g. if
	// dependency detection or specialised self-replacement
	// ordering ever lands).
	SelfReplacement bool

	// Optional flags a CR 614.10 "may" replacement — the owner
	// decides each time whether to apply it. When true, the apply-
	// loop queues a yes/no prompt (PendingChoiceOptionalReplacement)
	// before firing Replace. "Yes" → Replace runs normally; "No" →
	// the effect is marked applied without running Replace, and
	// the event proceeds unchanged (for this effect; other
	// mandatory replacements still fire). Used by CR 903.9
	// commander-zone replacement; future "may exile instead of
	// graveyard" cards would use it too. Added in S17 sub-PR 6.
	Optional bool

	// EntryLifeCost, when > 0, makes this a "you may pay N life; if
	// you don't, <replacement>" effect — the Ravnica shockland
	// cycle. The apply-loop queues a PendingChoiceEntryPayLife
	// prompt and bails; paying means Replace NEVER runs, declining
	// means it does. That inversion is why it isn't Optional: an
	// Optional "yes" applies the replacement, and here "yes" is what
	// avoids it, at a price.
	//
	// A player who can't legally pay (CR 118.4 — life total below
	// the cost) is not prompted; the replacement applies. Takes
	// precedence over Optional, which is meaningless alongside it.
	// See entry_choice.go. Added with the shockland cycle.
	EntryLifeCost int

	// CopySelector, when non-nil, makes this an "as this permanent
	// enters, you may have it enter as a copy of X" effect (CR
	// 707.2) — Clone, Phyrexian Metamorph, Spark Double, Sakashima
	// the Impostor. The apply-loop queues a
	// PendingChoiceCopyTarget picker and bails; the answer stamps
	// ev.EntersAsCopyOf and the entry path materialises it before
	// EventETB.
	//
	// Replace is never called for an effect with a selector: there
	// is nothing about the event left for it to rewrite, and the
	// selector's Except hook is where the card's "except" clause
	// goes. Takes precedence over Optional and EntryLifeCost, which
	// no printed copy effect combines with. See copy_choice.go.
	CopySelector *CopySelector

	// PureCancel declares that Replace does nothing but call
	// ev.Cancel() — it rewrites no other field on the event and
	// changes nothing else in the game.
	//
	// The engine uses it for exactly one thing. CR 616 asks the
	// affected player to order the applicable replacements, but when
	// EVERY applicable replacement is a pure cancel there is nothing
	// to order: whichever one is put first cancels the event, the
	// CR 616.1 apply-loop ends there, and the outcome is identical
	// for every permutation. Prompting would be a question with one
	// answer, so the engine applies them in gather order instead.
	//
	// Only the "skip this step" effects set it today — Stasis's untap
	// skip and the SkipYourDrawStep half of Necropotence and
	// Yawgmoth's Bargain — which is the case #710 was filed about:
	// two of them under one controller.
	//
	// Leaving it false is always safe; it costs a prompt, not
	// correctness. Setting it on an effect that does anything more is
	// NOT safe — the ordering it suppresses would have been
	// observable. Added in #710.
	PureCancel bool

	// PromptQuestion is the text rendered in the yes/no Optional
	// prompt. Short — fits in a modal header. Defaults to Label
	// when empty.
	PromptQuestion string

	// Label is the CR 616 prompt's header copy
	// ("Doubling Season: double counters"). Kept server-side so
	// the wire carries it; no localisation yet.
	Label string
}

// activeReplacement binds one declared ReplacementEffect to its
// source (nil for built-ins) plus the unique effect ID used for
// once-per-event tracking and the CR 616 prompt.
type activeReplacement struct {
	effect ReplacementEffect
	source *Card // nil for built-ins
	id     ReplacementEffectID
	// identity names WHICH DECLARED EFFECT this is an instance of,
	// as opposed to `id`, which names this instance of it. Zero for
	// an effect that has no stable identity — see
	// replacementIdentity.
	identity replacementIdentity
}

// replacementIdentity names one declared replacement effect: the
// catalog entry it is printed on, the slot it occupies in that
// entry's Replacements slice, and the player who controls the object
// contributing it.
//
// It is the engine's existing key for a catalog replacement with the
// object dropped. `ReplacementEffectID` is minted as
// `battlefieldIndex*256 + slot`, so it says "the Doubling Season in
// slot 3 of the battlefield"; this says "Doubling Season's counter-
// doubling replacement, controlled by Ian" — the same for every copy
// on the board. That is the whole point: two Rhox Faithmenders are
// two objects contributing ONE effect, and #792 is about not asking
// which of two identical modifications happened first.
//
// The zero value means "no stable identity, never interchangeable
// with anything". Built-in, turn-scoped and test replacements take
// it: they are registered per instance rather than declared on a
// catalog entry, so two of them are two separate declarations that
// happen to look alike, not two printings of one effect.
//
// The controller is part of the identity because a replacement
// routinely reads its own controller — Notion Thief's "you draw that
// card instead" writes `src.Controller` into the event, so two
// Notion Thieves under DIFFERENT controllers replace the same draw
// two different ways and the order between them is very much
// observable.
//
// What it does NOT capture: an effect whose Replace writes its own
// SOURCE into the event ("that damage is dealt to this creature
// instead", "put that counter on this creature instead"). Two copies
// of such a card would share an identity and would not be
// interchangeable. Nothing in the catalog does that today, and the
// note in AGENTS.md §7 tells card authors to flag it rather than
// ship it quietly. See sameModification.
type replacementIdentity struct {
	// card is the source's CatalogAbilityKey — the catalog entry
	// whose Replacements slice the effect came from. Empty for a
	// replacement that has no catalog entry behind it.
	card string
	// slot is the index into that entry's Replacements slice.
	slot int
	// controller is the controller of the object contributing the
	// effect — a replacement effect is a static ability, so its
	// controller is whoever controls its source.
	controller uuid.UUID
}

// known reports whether this identity is a real one. The zero value
// is "unidentified", which never matches anything — including
// another zero value.
func (r replacementIdentity) known() bool { return r.card != "" }

// errReplacementPending is a sentinel returned by
// applyReplacementsLocked when a CR 616 order-choose prompt was
// queued. The caller returns to the client so the prompt renders;
// ResolveReplacementOrder resumes the pipeline when the client
// submits the order.
//
// NOT a real error in the rules sense — it's a flow-control signal.
// Pipeline functions match on it with errors.Is.
var errReplacementPending = errors.New("replacement effect pending player choice")

// ErrReplacementIterationExceeded signals the CR 616.1 apply-loop
// hit its safety cap. Should never happen outside buggy cards; the
// engine emits EventEffectError and bails with this error so the
// caller can log or bubble it.
var ErrReplacementIterationExceeded = errors.New("replacement apply-loop iteration cap exceeded")

// maxReplacementIters is the safety cap on CR 616.1 iterations.
// Real games never exceed single-digit iterations (one per applied
// replacement). 32 is paranoid-generous.
const maxReplacementIters = 32

// RegisterReplacementForTest appends a ReplacementEffect to the
// per-game test-only slice. Test helper only — production code
// should populate effects.Spec.Replacements (catalog) or
// g.BuiltinReplacements (engine).
//
// The gather pass walks built-ins → catalog → test replacements in
// that order. Callers must hold g.mu (tests using the raw mutation
// API already hold it via the *Locked pattern).
func (g *Game) RegisterReplacementForTest(effect ReplacementEffect) {
	g.testReplacements = append(g.testReplacements, effect)
}

// applyReplacementsLocked runs the CR 614 replacement pipeline on
// ev. Returns:
//
//	out, nil        — event proceeded (possibly mutated). Caller
//	                  runs the underlying mutation using out's
//	                  payload.
//	nil, nil        — event canceled. Caller skips the underlying
//	                  mutation and the post-event EmitEvent.
//	ev, errReplacementPending
//	                — CR 616 order-choose prompt queued. Caller
//	                  returns to client; ResolveReplacementOrder
//	                  will resume.
//	ev, ErrReplacementIterationExceeded
//	                — apply-loop safety cap hit. Caller logs + lets
//	                  the event through as-is.
//
// Caller must hold g.mu. Mints ev.ID if zero. The once-per-event
// map entry for ev.ID is allocated here and cleared by defer at the
// pipeline function's outermost frame — NOT here — so CR 616 prompt
// pauses don't lose the map entry mid-iteration.
func (g *Game) applyReplacementsLocked(ev *ReplacementEvent) (*ReplacementEvent, error) {
	if ev == nil {
		return nil, nil
	}
	if ev.ID == 0 {
		ev.ID = ReplacementEventID(g.nextReplacementEventID.Add(1))
	}
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}

	for iter := 0; iter < maxReplacementIters; iter++ {
		if ev.Canceled {
			return nil, nil
		}
		applicable := g.gatherActiveReplacementsLocked(ev)
		if len(applicable) == 0 {
			return ev, nil
		}
		if len(applicable) > 1 {
			// CR 616: affected player picks order. Queue a prompt
			// and stash the resume frame; caller returns without
			// applying — with three exceptions, all of which apply
			// the gathered order inline instead.
			//
			// #710: every applicable effect is a PURE CANCEL, so
			// every ordering produces the same event. Two "skip your
			// draw step" enchantments under one controller are one
			// skipped draw step, and asking which of them skipped it
			// is a prompt with a single answer.
			if allPureCancels(applicable) {
				g.applyGatheredInOrderLocked(ev, applicable)
				continue
			}
			// #792: every applicable effect is the SAME declared
			// effect on two or more objects — two Doubling Seasons,
			// two Rhox Faithmenders. The player is being asked to
			// order N copies of one modification, so again there is
			// only one answer.
			if sameModification(applicable) {
				g.applyGatheredInOrderLocked(ev, applicable)
				continue
			}
			// Nobody is left to answer the prompt: the gathered
			// order stands (an eliminated player's prompt would
			// block the table forever; S31 fuzzer finding).
			if chooser := affectedPlayerForEvent(ev, applicable, g); g.chooserGoneLocked(chooser) {
				g.applyGatheredInOrderLocked(ev, applicable)
				continue
			}
			// #793: the event cannot pause — a life payment is a COST
			// (CR 118.3) and CR 601.2h pays a spell's costs as one
			// indivisible step. The gathered order stands, for the same
			// reason it does above: a prompt nobody can answer here is
			// a spell stuck on the stack with a half-paid cost.
			if ev.mustSettleNow {
				g.applyGatheredInOrderLocked(ev, g.skipQuestionsLocked(ev, applicable))
				continue
			}
			g.queueReplacementOrderPromptLocked(ev, applicable)
			return ev, errReplacementPending
		}
		// Exactly one applicable.
		chosen := applicable[0]
		if ev.mustSettleNow && asksItsOwnQuestion(chosen.effect) {
			// #793, the single-effect half of the branch above: an
			// effect that would ask its controller a question cannot
			// fire on an event that cannot pause, and firing it blind
			// would answer the question for them in the direction that
			// favours it. Skipped un-applied — weaker than printed,
			// never stronger, which is the posture
			// optionalReplacementResumableLocked takes for an entry
			// with nothing to resume it.
			g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
			continue
		}
		if chosen.effect.CopySelector != nil {
			// "You may have this enter as a copy of ..." — the
			// picker and its decline-inline cases live in
			// copy_choice.go. A queued prompt bails; anything else
			// has already marked the effect applied and falls
			// through to the next iteration.
			if g.offerCopyChoiceLocked(ev, chosen) {
				return ev, errReplacementPending
			}
			continue
		}
		if chosen.effect.EntryLifeCost > 0 {
			// "As this enters, you may pay N life." The prompt (and
			// the unaffordable-so-apply-it-inline case) lives in
			// entry_choice.go; a queued prompt bails, an inline
			// apply falls through to the next iteration.
			if g.offerEntryLifePaymentLocked(ev, chosen) {
				return ev, errReplacementPending
			}
			continue
		}
		if chosen.effect.Optional {
			// CR 614.10 "may" — owner decides each time. Queue a
			// yes/no prompt; the resume path either fires Replace
			// (yes) or marks applied and skips (no). A chooser who
			// has left the game gets the "no" inline: the commander
			// of an eliminated player going to the graveyard rather
			// than the command zone changes nothing for anyone.
			//
			// #359: so does an event with nothing to resume it. This
			// branch used to queue unconditionally while the
			// EntryLifeCost branch below has checked entryResumable
			// since #268, so an Optional self-replacement on a
			// permanent entering by an unresumable route would pause
			// with no way to finish — the card stranded in its old
			// zone and the prompt answerable to no effect. Taking the
			// un-applied branch is weaker than printed and never
			// stranded, which is the posture #268 chose and #272
			// reaffirmed. It is also what blocked the six reveal-
			// lands ("as this enters, you may reveal a land from your
			// hand") from shipping.
			if !g.optionalReplacementResumableLocked(ev) {
				g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
				continue
			}
			if g.chooserGoneLocked(g.optionalReplacementChooserLocked(ev, chosen)) {
				g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
				continue
			}
			g.queueOptionalReplacementPromptLocked(ev, chosen)
			return ev, errReplacementPending
		}
		// Mandatory: fire Replace inline and iterate.
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		if chosen.effect.Replace != nil {
			if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					ErrorMsg: err.Error(),
				})
			}
		}
	}
	// Iteration cap exceeded — bug. Emit diagnostic and let the
	// event through as-is so the game doesn't wedge.
	g.EmitEvent(Event{
		Kind:     EventEffectError,
		ErrorMsg: ErrReplacementIterationExceeded.Error(),
	})
	return ev, ErrReplacementIterationExceeded
}

// allPureCancels reports whether every gathered replacement has
// declared itself a pure cancel (see ReplacementEffect.PureCancel).
// When they all have, the CR 616 ordering prompt is unobservable —
// the first one applied cancels the event and ends the apply-loop,
// and they are interchangeable — so the engine skips the prompt.
//
// An effect that asks its controller a question first is never
// treated as a pure cancel however it is flagged — see
// asksItsOwnQuestion.
//
// An empty slice is not "all pure cancels": the only caller has
// already established len > 1, and answering true for nothing would
// be a trap for a future one.
func allPureCancels(applicable []activeReplacement) bool {
	if len(applicable) == 0 {
		return false
	}
	for _, a := range applicable {
		if !a.effect.PureCancel || asksItsOwnQuestion(a.effect) {
			return false
		}
	}
	return true
}

// sameModification reports whether every gathered replacement is an
// instance of ONE declared effect — the same slot of the same catalog
// entry under the same controller, on two or more objects. Two
// Doubling Seasons, two Rhox Faithmenders, two Hardened Scales.
//
// When they are, the CR 616 ordering prompt has one answer. The
// affected player is being asked to order N copies of a single
// modification, and doubling then doubling is doubling then doubling
// whichever Season is named first. Under #730's gate that prompt also
// holds the whole table until it is answered, so the click is not
// even free. #792.
//
// This deliberately does NOT extend to a MIXED window — two Doubling
// Seasons and a Hardened Scales. Collapsing the two Seasons there
// would force them to fire back to back, and CR 616.1 lets the
// affected player interleave: [DS, HS, DS] puts 6 counters on the
// creature and neither [DS, DS, HS] (5) nor [HS, DS, DS] (8) can
// reach it. So a window with any distinct effect in it prompts with
// every effect listed, exactly as before.
//
// An effect that asks its controller a question is excluded for the
// same reason allPureCancels excludes it — see asksItsOwnQuestion.
//
// Fewer than two is not "all the same": the only caller has already
// established len > 1, and answering true for a single effect (which
// never prompts anyway) would be a trap for a future one.
func sameModification(applicable []activeReplacement) bool {
	if len(applicable) < 2 {
		return false
	}
	want := applicable[0].identity
	if !want.known() {
		return false
	}
	for _, a := range applicable {
		if a.identity != want || asksItsOwnQuestion(a.effect) {
			return false
		}
	}
	return true
}

// asksItsOwnQuestion reports whether firing this effect puts a
// question to a player before it changes anything — a CR 614.10
// "may", a shockland's pay-life, a copy selector.
//
// Every caller here is deciding whether to skip the CR 616 ordering
// prompt and apply the gathered effects inline. Such an effect can
// never be part of that: skipping the ordering prompt would skip its
// question too, and firing Replace blind is the bug
// ResolveReplacementOrder's own CopySelector branch exists to avoid —
// the permanent enters as a 0/0 and nobody is asked anything. No
// printed card combines one of these with a reason to skip the
// prompt; the guard is here so a future one fails loudly by prompting
// rather than quietly by deciding.
func asksItsOwnQuestion(e ReplacementEffect) bool {
	return e.Optional || e.EntryLifeCost > 0 || e.CopySelector != nil
}

// applyGatheredInOrderLocked fires the gathered replacements in the
// order the gather pass produced, marking each applied, and stops at
// the first one that cancels the event. The un-prompted sibling of
// ResolveReplacementOrder's chosen-order loop, used when the CR 616
// prompt is skipped — because nobody could answer it (an eliminated
// chooser) or because every ordering gives the same answer
// (allPureCancels).
//
// Caller must hold g.mu, and must have allocated the once-per-event
// map entry for ev.ID.
func (g *Game) applyGatheredInOrderLocked(ev *ReplacementEvent, applicable []activeReplacement) {
	for _, chosen := range applicable {
		if ev.Canceled {
			break
		}
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		if chosen.effect.Replace != nil {
			if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
			}
		}
	}
}

// skipQuestionsLocked drops the gathered replacements that would ask
// their controller a question — a CR 614.10 "may", a shockland's
// pay-life, a copy selector — marking each applied so the apply-loop
// does not gather it again, and returns the rest in gather order.
//
// Only an event that cannot pause (mustSettleNow) uses it: on every
// other event the question is asked. Skipping is the weaker branch of
// a "may" and the un-copied branch of a selector, which is what "never
// stronger than printed" means here.
//
// Caller must hold g.mu, and must have allocated the once-per-event
// map entry for ev.ID.
func (g *Game) skipQuestionsLocked(ev *ReplacementEvent, applicable []activeReplacement) []activeReplacement {
	out := make([]activeReplacement, 0, len(applicable))
	for _, a := range applicable {
		if asksItsOwnQuestion(a.effect) {
			g.replacementsAppliedThisEvent[ev.ID][a.id] = true
			continue
		}
		out = append(out, a)
	}
	return out
}

// optionalReplacementResumableLocked reports whether pausing on ev
// for a CR 614.10 yes/no prompt has something that can finish the
// underlying mutation afterwards.
//
// Only a battlefield ENTRY can lack one. Every other event kind is
// completed inline by applyResolvedReplacementEventLocked from the
// event's own payload — a counter delta, a life change, a damage
// mark — and an EXIT move carries its route with it (zoneRoute, or
// the battlefield-leave path's own resume). An entry is different
// because finishing it generically can silently skip work the
// starting effect owed: an exile-return mints a new object identity
// (CR 400.7) and a library search owes its caller a shuffle, so those
// sites deliberately do not set entryResumable and must not pause.
//
// #359. Mirrors the guard offerEntryLifePaymentLocked has carried
// since #268.
//
// Caller must hold g.mu.
func (g *Game) optionalReplacementResumableLocked(ev *ReplacementEvent) bool {
	if ev == nil {
		return false
	}
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
		return true
	}
	if ev.OldZone == ZoneBattlefield {
		// Not an entry — a permanent staying put.
		return true
	}
	return ev.entryResumable
}

// clearReplacementEventLocked drops the per-event tracking map
// entry for ev.ID. Pipeline functions call this via defer at their
// outermost frame so CR 616 prompt pauses don't lose the entry
// mid-event. No-op when the entry is already gone.
//
// Caller must hold g.mu.
func (g *Game) clearReplacementEventLocked(id ReplacementEventID) {
	if g.replacementsAppliedThisEvent == nil {
		return
	}
	delete(g.replacementsAppliedThisEvent, id)
}

// gatherActiveReplacementsLocked walks built-ins + catalog +
// test replacements and returns all that apply to ev (pass Watches
// filter + AppliesTo predicate, not yet fired for this ev.ID).
//
// Ordering: built-ins first (commander zone), then battlefield
// cards in battlefield-slice order, then test replacements. Within
// a single card, slice-order. Ordering matters for the "exactly
// one applicable" short-circuit but not for the CR 616 prompt path
// (the prompt displays them for the chooser to reorder).
//
// Caller must hold g.mu.
func (g *Game) gatherActiveReplacementsLocked(ev *ReplacementEvent) []activeReplacement {
	if ev == nil {
		return nil
	}
	applied := g.replacementsAppliedThisEvent[ev.ID]
	var out []activeReplacement

	// Built-ins. IDs are minted as builtinReplacementIDBase + index.
	for i := range g.BuiltinReplacements {
		id := builtinReplacementIDBase + ReplacementEffectID(i)
		if applied[id] {
			continue
		}
		eff := g.BuiltinReplacements[i]
		if !eventKindMatches(eff.Watches, ev.Kind) {
			continue
		}
		if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, nil) {
			continue
		}
		out = append(out, activeReplacement{effect: eff, source: nil, id: id})
	}

	// Catalog — walk battlefield cards, look up each card's
	// replacements via CatalogReplacements, gather applicable ones.
	// IDs are minted as (battlefieldIndex * 256 + replacementIndex)
	// — small enough that catalog IDs never collide with the
	// built-in base range.
	if CatalogReplacements != nil {
		for cardIdx := range g.Battlefield.Cards {
			card := &g.Battlefield.Cards[cardIdx]
			// CatalogAbilityKey: a replacement effect is a static
			// ability (CR 614.1), so a permanent under a CR 613.1f
			// ability-removing effect contributes none.
			key := CatalogAbilityKey(*card)
			reps := CatalogReplacements(key)
			if len(reps) == 0 {
				continue
			}
			for repIdx := range reps {
				id := ReplacementEffectID(cardIdx*256 + repIdx + 1)
				if applied[id] {
					continue
				}
				eff := reps[repIdx]
				if !eventKindMatches(eff.Watches, ev.Kind) {
					continue
				}
				if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, card) {
					continue
				}
				out = append(out, activeReplacement{
					effect:   eff,
					source:   card,
					id:       id,
					identity: replacementIdentity{card: key, slot: repIdx, controller: card.Controller},
				})
			}
		}
	}

	// Self-replacements — a card replacing its OWN entry ("This land
	// enters tapped": every Temple, Guildgate and tri-land).
	//
	// These are invisible to the battlefield walk above, and that is
	// the whole reason this block exists: the replacement pipeline runs
	// PRE-push, so a card that is entering the battlefield is not on it
	// yet and cannot find its own effect. Worn Powerstone's OnETB
	// workaround — enter untapped, then tap — was the previous best
	// available, and it is observably different: the permanent really
	// does become tapped a beat after entering.
	//
	// Only consulted for a card that is NOT already on the
	// battlefield, so a permanent already in play can never match here
	// as well as in the walk above and apply the same effect twice.
	if CatalogReplacements != nil && ev.CardID != uuid.Nil && !g.Battlefield.Contains(ev.CardID) {
		const selfReplacementIDBase ReplacementEffectID = 1 << 45
		if entering, ok := g.LookupCardForEffect(ev.CardID); ok {
			key := CatalogKey(entering)
			reps := CatalogReplacements(key)
			for repIdx := range reps {
				id := selfReplacementIDBase + ReplacementEffectID(repIdx)
				if applied[id] {
					continue
				}
				eff := reps[repIdx]
				if !eventKindMatches(eff.Watches, ev.Kind) {
					continue
				}
				// `source` is the entering card itself, so an AppliesTo
				// comparing ev.CardID to source.InstanceID identifies
				// "this permanent" the same way it would on the
				// battlefield.
				src := entering
				if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, &src) {
					continue
				}
				out = append(out, activeReplacement{
					effect:   eff,
					source:   &src,
					id:       id,
					identity: replacementIdentity{card: key, slot: repIdx, controller: src.Controller},
				})
			}
		}
	}

	// Turn-scoped replacements — Fog-class effects registered via
	// RegisterTurnScopedReplacement. Live until StepCleanup clears
	// the slice. IDs in a dedicated range between catalog space
	// and test space.
	const turnScopedIDBase ReplacementEffectID = 1 << 50
	for i := range g.TurnScopedReplacements {
		id := turnScopedIDBase + ReplacementEffectID(i)
		if applied[id] {
			continue
		}
		eff := g.TurnScopedReplacements[i]
		if !eventKindMatches(eff.Watches, ev.Kind) {
			continue
		}
		if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, nil) {
			continue
		}
		out = append(out, activeReplacement{effect: eff, source: nil, id: id})
	}

	// Test replacements. IDs live in a dedicated range above the
	// catalog + turn-scoped spaces and below the built-in base.
	const testReplacementIDBase ReplacementEffectID = 1 << 55
	for i := range g.testReplacements {
		id := testReplacementIDBase + ReplacementEffectID(i)
		if applied[id] {
			continue
		}
		eff := g.testReplacements[i]
		if !eventKindMatches(eff.Watches, ev.Kind) {
			continue
		}
		if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, nil) {
			continue
		}
		out = append(out, activeReplacement{effect: eff, source: nil, id: id})
	}

	return out
}

// RegisterTurnScopedReplacement appends a replacement effect that
// lives until the current turn's StepCleanup. Used by Fog and
// similar "until end of turn" prevention cards. Caller must hold
// g.mu.
func (g *Game) RegisterTurnScopedReplacement(effect ReplacementEffect) {
	g.TurnScopedReplacements = append(g.TurnScopedReplacements, effect)
}

// ClearTurnScopedReplacementsLocked drops every turn-scoped
// replacement at StepCleanup. Called from runStepEntryHooksLocked.
// Caller must hold g.mu.
func (g *Game) ClearTurnScopedReplacementsLocked() {
	if len(g.TurnScopedReplacements) > 0 {
		g.TurnScopedReplacements = nil
	}
}

// ReplacementOptionMetaForEffect returns the prompt label and
// source card ID for the given ReplacementEffectID. Used by the
// protocol view layer to stamp ReplacementOptionView entries.
// Returns ("", uuid.Nil) when the ID doesn't resolve (stale
// prompts, mis-assigned IDs). Walks the built-ins + catalog + test
// registries in the same order as gatherActiveReplacementsLocked,
// using the same ID minting scheme.
//
// Caller must hold g.mu (read lock is sufficient — no mutation).
func (g *Game) ReplacementOptionMetaForEffect(id ReplacementEffectID) (string, uuid.UUID) {
	// Built-ins.
	if id >= builtinReplacementIDBase {
		i := int(id - builtinReplacementIDBase)
		if i >= 0 && i < len(g.BuiltinReplacements) {
			return g.BuiltinReplacements[i].Label, uuid.UUID{}
		}
		return "", uuid.UUID{}
	}
	const testReplacementIDBase ReplacementEffectID = 1 << 55
	const turnScopedIDBase ReplacementEffectID = 1 << 50
	// Test replacements.
	if id >= testReplacementIDBase {
		i := int(id - testReplacementIDBase)
		if i >= 0 && i < len(g.testReplacements) {
			return g.testReplacements[i].Label, uuid.UUID{}
		}
		return "", uuid.UUID{}
	}
	// Turn-scoped replacements (Fog etc.).
	if id >= turnScopedIDBase {
		i := int(id - turnScopedIDBase)
		if i >= 0 && i < len(g.TurnScopedReplacements) {
			return g.TurnScopedReplacements[i].Label, uuid.UUID{}
		}
		return "", uuid.UUID{}
	}
	// Catalog. ID = cardIdx*256 + repIdx + 1, so decode accordingly.
	if CatalogReplacements == nil {
		return "", uuid.UUID{}
	}
	raw := int(id) - 1
	cardIdx := raw / 256
	repIdx := raw % 256
	if cardIdx < 0 || cardIdx >= len(g.Battlefield.Cards) {
		return "", uuid.UUID{}
	}
	card := &g.Battlefield.Cards[cardIdx]
	reps := CatalogReplacements(CatalogAbilityKey(*card))
	if repIdx < 0 || repIdx >= len(reps) {
		return "", card.InstanceID
	}
	return reps[repIdx].Label, card.InstanceID
}

// ReplacementEffectIDFromString parses a decimal string ID back
// into a ReplacementEffectID. Mirror of the fmt used by the view
// layer's replacementEffectIDString. Used by the actions
// dispatcher to decode the client's `order []string` payload.
func ReplacementEffectIDFromString(s string) (ReplacementEffectID, error) {
	var v uint64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, errReplacementIDMalformed
		}
		v = v*10 + uint64(r-'0')
	}
	return ReplacementEffectID(v), nil
}

// ReplacementEffectIDToString is the canonical string form used on
// the wire. Decimal digits, no prefix. Matches
// ReplacementEffectIDFromString's grammar.
func ReplacementEffectIDToString(id ReplacementEffectID) string {
	if id == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = byte('0' + id%10)
		id /= 10
	}
	return string(buf[i:])
}

var errReplacementIDMalformed = errors.New("replacement effect id malformed")

// eventKindMatches reports whether a ReplacementEventKind is
// compatible with one of the EventKind values in watches. The
// mapping is deliberate (a replacement watching "zone_move" fires
// for RepEventMove; one watching "counter_placed" fires for
// RepEventCounter; etc.). Empty watches matches any kind.
func eventKindMatches(watches []EventKind, kind ReplacementEventKind) bool {
	if len(watches) == 0 {
		return true
	}
	var want EventKind
	switch kind {
	case RepEventDraw:
		want = EventDrawCard
	case RepEventMove:
		want = EventZoneMove
	case RepEventCounter:
		want = EventCounterPlaced
	case RepEventLife:
		want = EventChangeLife
	case RepEventDamage:
		want = EventDealDamage
	case RepEventStepTransition:
		// No corresponding public EventKind — step transitions
		// don't emit Event today. Watches match if the slice
		// contains the engine-only sentinel EventStepTransition,
		// which replacements.go exports below.
		want = EventStepTransition
	default:
		return false
	}
	for _, w := range watches {
		if w == want {
			return true
		}
	}
	return false
}

// chooserGoneLocked reports whether a prompt addressed to id could
// never be answered: no such seat, or the seat has been eliminated.
// Caller must hold g.mu.
func (g *Game) chooserGoneLocked(id uuid.UUID) bool {
	if id == uuid.Nil {
		return false
	}
	p := g.playerByIDLocked(id)
	return p == nil || p.Eliminated
}
