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
//   - drawCardLocked → RepEventDraw (carries DrawCount, #1222)
//   - produceManaLocked → RepEventProduceMana (CR 106.12b, #1222;
//     produce_mana.go, reached from all four production sites)
//   - MoveCardByIDAsCommander → RepEventMove (carries EntersTapped,
//     EntersWithCounters, asCommanderMove breadcrumb)
//   - routeCardToZoneLocked → RepEventMove, or RepEventDiscard when the
//     route is a discard (CR 701.8, #650)
//   - createTokensLocked → RepEventCreateTokens, then one
//     RepEventMove battlefield entry per token created (CR 701.7b,
//     #762)
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
//	"draw"         — RepEventDraw    — DrawPlayer, DrawCount (CR 121.2)
//	"produce_mana" — RepEventProduceMana — ManaPlayer, ManaSource, ManaColors (CR 106.12b)
//	"move"         — RepEventMove    — CardID, OldZone, NewZone, NewZoneOwner, EntersTapped, EntersAttacking, EntersWithCounters, asCommanderMove
//	"discard"      — RepEventDiscard — CardID, DiscardPlayer, DiscardCause, NewZone, NewZoneOwner (CR 701.8)
//	"counter"      — RepEventCounter — CounterTarget OR CounterPlayer,
//	                 CounterName, CounterDelta, CounterPlacer,
//	                 CounterFromCombatDamage
//	"life"         — RepEventLife    — LifePlayer, LifeDelta
//	"damage"       — RepEventDamage  — DamageSource, DamageTarget, DamageAmount, IsCombatDamage
//	"create_tokens"— RepEventCreateTokens — TokenController, TokenGroups, TokenAttacking (CR 701.7b)
//	"keyword_action"— RepEventKeywordAction — KeywordAction, KeywordActionCount (CR 701.22 / 701.25 / 701.34)
//	"mill"         — RepEventMill    — MillPlayer, MillCount (CR 701.13a)
//	"step"         — RepEventStepTransition — StepTransitionStep, StepTransitionSeat
type ReplacementEventKind string

const (
	RepEventDraw    ReplacementEventKind = "draw"
	RepEventMove    ReplacementEventKind = "move"
	RepEventCounter ReplacementEventKind = "counter"
	RepEventLife    ReplacementEventKind = "life"
	RepEventDamage  ReplacementEventKind = "damage"

	// RepEventDiscard is a discard (CR 701.8a) — the one exit whose
	// keyword action is defined by where the card comes FROM, so it is
	// its own event kind rather than a flag on RepEventMove. Library of
	// Leng, madness (#657) and the Obstinate Baloth family all key on
	// "if you would discard", and two of the three also need the CAUSE:
	// an effect's instruction, a cost, or the cleanup step's turn-based
	// action. See #650 and ADR 0061.
	//
	// It still carries the move payload — CardID, OldZone (always the
	// hand), NewZone, NewZoneOwner and the zoneRoute — because a discard
	// IS a move out of the hand and its replacements rewrite the
	// destination. Everything in the engine that finishes, abandons or
	// prunes a routed exit therefore handles the two kinds together; see
	// isExitMove.
	RepEventDiscard ReplacementEventKind = "discard"

	// RepEventCreateTokens is one "create N tokens" INSTRUCTION
	// (CR 701.7b), opened once however many tokens it makes, so a
	// doubler modifies the instruction rather than each token: "create
	// two tokens" with Parallel Lives out is one event that becomes
	// four tokens, not two events of two. See #762 and ADR 0061.
	//
	// The tokens it settles on each run the ordinary battlefield-ENTRY
	// pipeline afterwards (a RepEventMove with no old zone), so
	// enters-tapped, enters-with-counters and the ETB hook reach a
	// token exactly as they reach every other permanent.
	RepEventCreateTokens ReplacementEventKind = "create_tokens"

	// RepEventKeywordAction is one KEYWORD ACTION with a count —
	// proliferate (CR 701.34), scry (CR 701.22) or surveil
	// (CR 701.25) — opened once per INSTRUCTION at the one entry point
	// of each, the way RepEventCreateTokens is opened once per
	// creation instruction. "If you would proliferate, proliferate
	// twice instead" (Tekuthal, Inquiry Dominus) and "if you would
	// scry, scry that many plus one instead" are replacements of the
	// action, not of the counters it places or the cards it looks at.
	// See #976 and keyword_action.go.
	//
	// ONE kind with an Action discriminator rather than one kind per
	// verb: everything downstream of the count is the same code for
	// all three, so a fourth counted keyword action is a constant and
	// an arm of one switch. What the count MEANS is per action —
	// times for proliferate, cards for scry and surveil — and is
	// written on the KeywordAction constants.
	RepEventKeywordAction ReplacementEventKind = "keyword_action"

	// RepEventMill is one "mill N cards" INSTRUCTION (CR 701.13a),
	// opened once per instruction before any card moves — the third
	// member of the count-carrying family, after RepEventCreateTokens
	// and RepEventKeywordAction, and opened for the same reason: "if an
	// opponent would mill one or more cards, they mill twice that many
	// cards instead" (Bruvac the Grandiloquent) replaces the NUMBER,
	// not the per-card zone change.
	//
	// The per-card zone change is a separate, older window and was
	// never missing: every milled card routes through
	// routeCardToZoneLocked, so Leyline of the Void sees each card and
	// a milled commander is offered the command zone (CR 903.9). What
	// had no seam was the amount. See #569 and mill.go.
	//
	// Opened only for a real mill: a graveyard destination (CR 701.13a
	// defines the keyword action by where the cards go, so "exile the
	// top N cards of your library" is not a mill and opens no window)
	// and a positive count (there is nothing to replace about milling
	// nothing, and the unbounded `until` run — Helm of Obedience — names
	// no number to double).
	RepEventMill ReplacementEventKind = "mill"

	// RepEventProduceMana is one MANA PRODUCTION (CR 106.12b), opened
	// once per batch of mana an object is about to put into a pool and
	// before any of it is there — the fourth member of the
	// count-carrying family after RepEventCreateTokens,
	// RepEventKeywordAction and RepEventMill, and opened for the same
	// reason: "if a permanent you control would produce one or more
	// mana, it produces twice as much of that mana instead" (Mana
	// Reflection, Nyxbloom Ancient) replaces the AMOUNT, not the
	// spending of it.
	//
	// It carries COLOURS rather than a bare count, because "twice as
	// much of THAT mana" names the mana that was going to be produced:
	// a Forest doubles into {G}{G} and a Sol Ring into {C}{C}{C}{C}.
	// The colour has to be settled before the window opens, which is
	// why a multi-option slot (Birds of Paradise) opens it from
	// ResolveManaChoice — the one moment the produced colour is known
	// — and not at the activation. See produce_mana.go and ADR 0013
	// §5ab.
	//
	// It is the one kind that can NEVER pause. CR 605.3a makes
	// activating a mana ability one indivisible step with no stack and
	// no priority window inside it, and the auto-tapper's contract is
	// "no further player decisions", so every event of this kind sets
	// mustSettleNow and the CR 616.1 ordering prompt is applied in
	// gather order instead. That is a declared simplification and it is
	// unobservable for every printed card on the row: both of them
	// multiply, and multiplication commutes.
	RepEventProduceMana ReplacementEventKind = "produce_mana"

	RepEventStepTransition ReplacementEventKind = "step"
)

// isExitMove reports whether a kind is a routed move OUT of a zone — an
// ordinary RepEventMove or a discard. The two share every field the
// exit path reads (CardID, OldZone, NewZone, NewZoneOwner, zoneRoute)
// and differ only in what the completed move announces, so every engine
// site that finishes, abandons or prunes a routed exit asks this rather
// than naming one kind and quietly skipping the other.
func isExitMove(kind ReplacementEventKind) bool {
	return kind == RepEventMove || kind == RepEventDiscard
}

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

// The ReplacementEffectID space, low to high. Each registry the
// gather pass walks gets a disjoint range, so an ID names both the
// registry it came from and the entry within it. The catalog's range
// is the bottom one and is the only one with two coordinates packed
// into it — see encodeCatalogReplacementID.
//
//	1                     .. selfReplacementIDBase  catalog (battlefield card × slot)
//	selfReplacementIDBase .. turnScopedIDBase       a card replacing its own entry, by slot
//	turnScopedIDBase      .. testReplacementIDBase  turn-scoped (Fog), by index
//	testReplacementIDBase .. builtinReplacementIDBase  test-injected, by index
//	builtinReplacementIDBase ..                     built-ins (commander zone), by index
//
// They were three `const` declarations inside the two functions that
// read them, declared twice over; #801 moved them here so the scheme
// is written down once.
const (
	selfReplacementIDBase    ReplacementEffectID = 1 << 45
	turnScopedIDBase         ReplacementEffectID = 1 << 50
	testReplacementIDBase    ReplacementEffectID = 1 << 55
	builtinReplacementIDBase ReplacementEffectID = 1 << 62
)

// MaxCatalogReplacementSlots is the number of Replacements one
// catalog entry may declare. It is the stride of the catalog ID
// encoding (see encodeCatalogReplacementID), so a card that declared
// more would mint IDs belonging to the next card along and silently
// replace the wrong permanent's effect in a CR 616 prompt.
//
// effects.Register rejects a Spec over the budget at boot, the way it
// rejects every other unrepresentable declaration. No printed card is
// anywhere near it: the fullest entry in the catalog today (Uncivil
// Unrest) declares two.
const MaxCatalogReplacementSlots = 256

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

	// DrawCount is HOW MANY cards this draw instruction draws, and the
	// one field a draw-amount replacement rewrites: Thought Reflection
	// and Alhammarret's Archive are `ev.DrawCount *= 2`.
	//
	// The base is always ONE. CR 121.2 makes "draw three cards" three
	// individual card draws, so DrawNForEffect loops and each
	// repetition opens its own window — which is what makes a doubler
	// double EACH of them (a Thought Reflection on "draw three" draws
	// six) and what makes a redirect (Notion Thief) able to take one
	// card of three rather than all or nothing.
	//
	// What the settled count does NOT do is re-open the window. The N
	// cards of a replaced draw are performed one at a time
	// (actuallyDrawCardsLocked, so every per-card payoff — Nekusar,
	// Sheoldred, Consecrated Sphinx — still fires per card) but they
	// are ONE event, CR 614.5's once-per-event tracking covering all
	// of them. Two Thought Reflections therefore draw FOUR, not three:
	// the second one multiplies the count the first left, through the
	// ordinary CR 616.1 apply-loop, exactly as two mill doublers do.
	//
	// A hand-built event that leaves it zero draws one card, because
	// the honest spelling of "no draw" is ev.Cancel() and a silent
	// zero would be a draw that vanished. See ADR 0013 §5ab.
	DrawCount int

	// --- RepEventProduceMana fields ---

	// ManaPlayer is the player whose pool the mana is about to reach —
	// the controller of the ability, or the player a resolving spell
	// adds it for. Actor carries the same value; this is the name the
	// rules use, and it is the CR 616.1 affected player.
	ManaPlayer uuid.UUID

	// ManaSource is the OBJECT producing the mana (CR 106.12b names
	// the object, not the ability): the permanent whose mana ability
	// was activated, or the spell that said "Add {B}{B}{B}". Source
	// carries the same value; this is the name a replacement's
	// AppliesTo reads, because "if a PERMANENT you control would
	// produce" is a question about the object and is false for a
	// spell.
	ManaSource uuid.UUID

	// ManaColors is the mana about to be produced, one entry per mana,
	// in the order it would reach the pool — and the one field a
	// mana-production replacement rewrites. "Twice as much of that
	// mana" is MultiplyMana(2), which repeats each entry.
	//
	// Colours rather than a count because the printed cards say "twice
	// as much of THAT mana": a doubled Forest is {G}{G} and a doubled
	// Sol Ring {C}{C}{C}{C}, and a count alone could not say which.
	// Every entry is a settled colour — a pipe slot has already been
	// answered by the time the window opens, which is why the pick
	// path opens it and the activation does not.
	//
	// An empty list after a replacement is the same outcome as
	// ev.Cancel(): no mana is produced.
	ManaColors []string

	// ManaFromTap says this production is part of TAPPING A PERMANENT
	// FOR MANA (CR 106.12a) — a mana ability with a {T} in its cost —
	// as opposed to a spell's "Add {B}{B}{B}", a mana ability with no
	// tap, or a triggered mana ability's own output.
	//
	// It is the printed condition on both cards this event was built
	// for: Mana Reflection and Nyxbloom Ancient say "if you TAP a
	// permanent for mana, it produces twice/three times as much of
	// that mana instead", so a Dark Ritual is not doubled and neither
	// is Wild Growth's extra {G} — the enchantment was not tapped.
	//
	// The same bit PendingChoice.ManaTapped carries for the triggered
	// mana abilities (ADR 0074 §3), read off the same two places and
	// for the same reason: it is a fact about how the mana came to be
	// produced, and nothing downstream could reconstruct it.
	ManaFromTap bool

	// What is deliberately NOT on this event is the SPEND RESTRICTION
	// the minted tokens carry (Eldrazi Temple's "colorless Eldrazi
	// only", Ancient Ziggurat's creature spells). It is a parameter of
	// produceManaLocked instead, because CR 106.12b replaces HOW MUCH
	// mana is produced and nothing in the rules lets a replacement
	// change what the produced mana may be spent on — so putting it
	// here would be offering the catalog a rewrite the rules do not
	// have. The replaced tokens still carry it, which is what keeps a
	// doubled Ziggurat from putting unrestricted mana in the pool
	// (#259).

	// --- RepEventMove fields ---

	// CardID is the card being moved.
	CardID uuid.UUID

	// OldZone / NewZone / NewZoneOwner describe the motion.
	// Replacements can rewrite NewZone (the CR 903.9 commander-zone
	// built-in, Stone of Erech's graveyard → exile, Library of Leng's
	// hand → top of library).
	//
	// A created TOKEN enters with OldZone empty: it came from no zone
	// at all (CR 111.1 — a token is created on the battlefield), which
	// is the honest spelling and the one every "enters the
	// battlefield" replacement already reads past, because they key on
	// NewZone.
	OldZone      ZoneKind
	NewZone      ZoneKind
	NewZoneOwner uuid.UUID

	// Destruction says this battlefield exit is a DESTRUCTION
	// (CR 701.7a), as opposed to the other things that take the same
	// exit — a sacrifice (CR 701.21a), the legend rule, an illegally
	// attached Aura, a creature at zero toughness, a planeswalker at
	// zero loyalty, a battle at zero defense. Only meaningful on a
	// RepEventMove whose OldZone is the battlefield.
	//
	// DECLARED, never derived. Every one of those exits ends in the
	// same graveyard through the same primitive, so there is nothing
	// about the event a reader could have looked at to tell them
	// apart; the engine had no use for the distinction until
	// regeneration needed it (#667), and indestructible sidesteps it
	// by filtering BEFORE the window opens (indestructible.go).
	//
	// Set from the route the destroying verb chose (destroyRoute in
	// simultaneous.go), so it survives a CR 903.9 pause with
	// everything else the move was asked for.
	Destruction bool

	// CantBeRegenerated is the rider printed by Damnation, Day of
	// Judgment, Mortify, Putrefy, Pongify and the rest: this
	// destruction ignores regeneration shields (CR 701.19c).
	//
	// It gates the built-in's AppliesTo rather than being consumed
	// inside its Replace, because CR 701.19d leaves an ignored shield
	// UNUSED — a creature with a shield that Damnation kills would
	// still have had that shield if something had saved it. Only
	// meaningful alongside Destruction.
	CantBeRegenerated bool

	// EntersTapped is mutated by enters-tapped replacements
	// (Kismet). Only meaningful when NewZone == ZoneBattlefield.
	// The battlefield-entry path reads this and sets Card.Tapped
	// before emitting EventETB.
	EntersTapped bool

	// EntersPrepared is CR 722.3a's "this creature enters prepared",
	// a CR 614.1d replacement (ADR 0090): set by the entering card's
	// own self-replacement (effects.SelfEntersPrepared) and read by
	// every battlefield landing, which gives the permanent the
	// designation — and makes its CR 722.3c copy in exile — before
	// EventETB, so the permanent is never on the battlefield
	// unprepared. Rides the event for EntersTapped's reason. Only
	// meaningful when NewZone == ZoneBattlefield, and ignored for a
	// permanent with no prepare spell (CR 722.3a).
	EntersPrepared bool

	// EntersAttacking is the player, planeswalker or battle the
	// permanent is put onto the battlefield ATTACKING (CR 506.3c,
	// #1227) — ninjutsu's "put this card onto the battlefield from
	// your hand tapped and attacking" (CR 702.49a) and the attack a
	// token is created making (Parhelion II). uuid.Nil, which is every
	// other entry in the game, enters not attacking.
	//
	// It rides the EVENT rather than a local for the reason
	// EntersTapped and FaceDown do: both battlefield-entry doors carry
	// the same fact in the same field, a replacement effect inspecting
	// the entry sees it, and a CR 616 resume that only has the event
	// still knows what the entry was for. Both doors hand it to
	// stampEntryAttackerLocked (attackers.go), which is the one place
	// CR 506.3c's "never declared" marking happens.
	//
	// Only meaningful when NewZone == ZoneBattlefield. Nothing
	// REPLACES it today; it is seeded by the caller and read by the
	// entry.
	EntersAttacking uuid.UUID

	// FaceDown is the CR 708.2 state the permanent ENTERS in —
	// FaceDownManifested for a manifest (CR 701.34a), FaceDownMorphed
	// or FaceDownDisguised for a spell that was cast face down
	// (CR 708.4). The zero value is an ordinary face-up entry, which
	// is every other entry in the game.
	//
	// Seeded by the caller that knows (stack resolution off
	// StackItem.FaceDown, putOntoBattlefieldFromZoneLocked off
	// ZoneEntryOptions.FaceDown) and read by the ONE entry finisher,
	// so the fact rides the same event for both doors and a
	// replacement effect inspecting the entry sees the same truth
	// whichever one the permanent came through.
	//
	// It changes two things about the landing, both of them CR 708
	// falling out rather than being special-cased: the entry sets the
	// face-down state and its CR 708.5 viewers instead of marking
	// every seat a knower, and because the state is set before the
	// announcement CatalogKey answers "" — so a face-down entry runs
	// no ETB trigger and no "as enters" choice (CR 708.2a). Only
	// meaningful when NewZone == ZoneBattlefield. Added for #1194
	// (ADR 0082 decision 3).
	FaceDown FaceDownKind

	// FaceDownListed is the CR 708.2 body the putting effect LISTED
	// for a face-down entry — Yedora's "It's a Forest land",
	// Cybership's "They're 2/2 Cyberman artifact creatures". nil is
	// CR 708.2a's default nameless 2/2. Rides the event beside
	// FaceDown for FaceDown's reason: the one entry finisher reads it,
	// whichever door the permanent came through, and a paused entry's
	// resume still knows it. Never mutated through the pointer, so the
	// undo clone may share it. Meaningless without FaceDown. #1270.
	FaceDownListed *FaceDownListing

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
	//
	// Write it through AddCounterAtETB and DRAIN IT THROUGH
	// (*Game).applyEntryCountersLocked — never with a bare range
	// (#1010). Each kind opens its own RepEventCounter window, so the
	// drain order is observable: map order made two kinds on one entry
	// open their windows, ask their CR 616 prompts and write their log
	// lines differently on every run. The drain's doc comment carries
	// the canonical order and why it is not a player's choice.
	EntersWithCounters map[string]int

	// entryResumable is an unexported breadcrumb meaning "if this
	// entry pauses for a prompt, the generic resume path may finish
	// it" (executeEntryToBattlefieldLocked). Set by the land-play
	// branch of CastSpell, by stack resolution — since S16.5, so Clone
	// can ask what to copy — which hands the resume its StackItem so
	// the Aura attach and evoke's sacrifice trigger survive the pause,
	// and since #478 by the three effect-side entries that go through
	// enterBattlefieldThroughPipelineLocked: the library search, the
	// exile return and the reanimation.
	//
	// Those three used to be unflagged, on the argument that finishing
	// them generically would silently drop work the starting effect
	// owed — the search's shuffle and EventSearchLibrary, the caller's
	// Then, and the exile return's new object identity (CR 400.7). The
	// argument was right about the cost; what was missing was a way to
	// CARRY those duties across the pause, which is what entryTail
	// below is. With one the generic resume is faithful, so they are
	// flagged and the entry can ask its question.
	//
	// Since #1322 the hand / library / exile "put onto the battlefield"
	// batch is flagged too (entry_batch.go). It could not be while its
	// only resume was the single-card finisher, which would land one
	// card after its siblings had been announced; its events carry the
	// batch on entryTail instead, and the resume hands each settled
	// event back to the batch rather than landing it. The one entry
	// left unflagged is the sandbox move_card verb
	// (moveCardByRefLocked), a manual move with nothing behind it to
	// finish, whose questions take their un-asked branch.
	//
	// An effect that WOULD pause consults this before prompting: a
	// pay-life entry choice on an unflagged event takes the un-paid
	// branch rather than stranding the card in its old zone (see
	// offerEntryLifePaymentLocked). Any entry site that grows a
	// faithful resume should set this and inherit the prompt.
	entryResumable bool

	// entryTail is the entry half's answer to zoneRoute: what the
	// effect that asked for this battlefield ENTRY still owes once the
	// pipeline settles — the new object identity an exile return mints,
	// and the rest of the effect (a search's shuffle and its caller's
	// Then). Set by enterBattlefieldThroughPipelineLocked's callers and
	// read only by executeEntryToBattlefieldLocked and
	// runEntryTailLocked, which both the inline path and the CR 616
	// resume go through.
	//
	// A non-nil value is not what makes an entry resumable —
	// entryResumable is — but it is what makes resuming it FAITHFUL.
	// See entry_tail.go. Added in #478.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	entryTail *entryTail

	// landPlay flags the one battlefield entry that is a LAND DROP
	// (CR 305.2): the land branch of CastSpell. The resume bumps
	// LandsPlayedThisTurn only for it.
	//
	// It is declared rather than derived. The resume used to infer a
	// land play from "this card is a land and the event carries no
	// StackItem", which was true while the only two resumable entries
	// were the land play and stack resolution and became wrong the
	// moment a fetched, reanimated or blinked land could pause there
	// (#478): a land an effect puts onto the battlefield was not
	// PLAYED, so it must not spend the turn's land drop.
	landPlay bool

	// zoneRoute is the exit half's answer to entryResumable: the
	// per-destination bookkeeping (to the bottom of the library, face
	// down in exile, this was a mill, this was a counterspell, this
	// was a discard) that a paused move has to carry across the pause
	// so the resume can finish it exactly as the mover asked — plus,
	// since #853, the rest of the effect that asked for it. Set by
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

	// --- RepEventDiscard fields ---
	//
	// A discard also fills in the RepEventMove fields above: CardID is
	// the discarded card, OldZone is the hand it is leaving, and
	// NewZone / NewZoneOwner are the destination a replacement may
	// rewrite.

	// DiscardPlayer is the player discarding the card — its owner,
	// because every hand in this engine holds only its owner's cards
	// (CR 701.8a moves the card to that player's graveyard).
	DiscardPlayer uuid.UUID

	// DiscardCause is why the discard is happening: an effect's
	// instruction, a cost, or the cleanup step's turn-based action. It
	// is the distinction the rules actually draw, and the one Library
	// of Leng needs ("if an EFFECT causes you to discard a card"); the
	// Gatherer ruling of 2004-10-04 says costs are not effects, and
	// CR 514.1 / 703.1 make the cleanup discard a turn-based action
	// rather than anybody's effect. See ADR 0013 §10a, which withdrew
	// the voluntary/involuntary framing #160 was written around, and
	// ADR 0061.
	//
	// Source names the card whose effect or cost asked for it, so the
	// Obstinate Baloth family ("a spell or ability an opponent controls
	// causes you to discard") can read its controller.
	DiscardCause DiscardCause

	// --- RepEventCreateTokens fields ---

	// TokenController is the player the tokens are created under the
	// control of. Actor carries the same value; this is the name the
	// rules use ("create one or more tokens UNDER YOUR CONTROL"), and
	// a doubler's AppliesTo reads it.
	TokenController uuid.UUID

	// TokenGroups is what the instruction creates, one entry per KIND
	// of token: a template, how many of it, and the creation-time
	// entry options the instruction asked for.
	//
	// Groups rather than a bare count because the printed cards need
	// both halves. Parallel Lives, Anointed Procession, Doubling
	// Season, Primal Vigor and Mondrak multiply the COUNTS
	// (MultiplyTokens); Academy Manufactor rewrites the KIND SET
	// (ReplaceTokenKinds) — "if you would create a Clue, Food or
	// Treasure token, instead create one of each" turns one group into
	// three. A count alone could not express the second.
	TokenGroups []TokenGroup

	// TokenAttacking is the player the tokens are created attacking
	// (CR 506.3c — put onto the battlefield attacking, never
	// declared, so nothing sees an attack declaration). uuid.Nil for
	// the ordinary creation.
	TokenAttacking uuid.UUID

	// tokenTail is the token half's answer to lifeTail and damageTail:
	// the rest of the effect that asked for the creation, run with the
	// IDs of the tokens that were actually made once the pipeline
	// settles. Set by CreateTokensThenForEffect and read only by
	// runTokenTailLocked, which both the inline path and the CR 616
	// resume reach.
	//
	// It exists because a creation can now PAUSE: a Doubling Season
	// and an Academy Manufactor in the same window is a CR 616
	// ordering prompt, and a caller that says "create a token, then
	// sacrifice it" cannot write the second half on the next line.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	// See token_create.go.
	tokenTail *tokenTail

	// --- RepEventKeywordAction fields ---

	// KeywordAction is WHICH keyword action this is — proliferate,
	// scry or surveil. A replacement narrows on it in AppliesTo, the
	// way a discard replacement narrows on DiscardCause; Actor is the
	// player taking the action ("if YOU would proliferate") and
	// Source the card whose effect asked for it.
	KeywordAction KeywordAction

	// KeywordActionCount is the count the action is taken with, and
	// the one field a keyword-action replacement rewrites: "twice
	// instead" is `ev.KeywordActionCount *= 2`, "that many plus one"
	// is `+= 1`.
	//
	// What it counts is per action and is written on the
	// KeywordAction constants: for proliferate it is the number of
	// TIMES the whole action is taken (base 1), for scry and surveil
	// the number of CARDS looked at (base N).
	KeywordActionCount int

	// keywordAction is the keyword-action sibling of zoneRoute,
	// damageTail and tokenTail: what the entry point still owes once
	// the window settles — a proliferate's chosen permanents and
	// players, and a scry's prompt kind and "then" continuation. Set
	// by the entry points in keyword_action.go and read only by
	// applyResolvedKeywordActionLocked, which both the inline path and
	// the CR 616 resume go through.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	keywordAction *keywordActionTail

	// --- RepEventMill fields ---

	// MillPlayer is the player milling — the one whose library is read,
	// which is NOT always the card's owner or the effect's controller
	// ("each opponent mills three cards"). Actor carries the same
	// value; this is the name the rules use, and it is the affected
	// player for CR 616.1.
	MillPlayer uuid.UUID

	// MillCount is how many cards the instruction mills, and the one
	// field a mill replacement rewrites: Bruvac the Grandiloquent is
	// `ev.MillCount *= 2`, The Water Crystal is `+= 4`.
	//
	// It is the number the instruction ASKED for, not what the library
	// can supply. A player told to mill more cards than they have mills
	// as many as possible (CR 701.13b) and the clamp happens in
	// millPlanLocked, after this window settles — so a Bruvac doubling
	// a mill of twenty against a library of twelve doubles twenty,
	// which is what the card says and is observable through any
	// "plus N" sharing the window.
	MillCount int

	// mill is the mill's tail: what the instruction still owes once the
	// window settles — the destination it named and the caller's
	// continuation. Set by the two entry points in effect_api.go and
	// read only by applyResolvedMillLocked, which both the inline path
	// and the CR 616 resume go through.
	//
	// #1176: no `until` rides here. A run is a SEQUENCE of one-card
	// mill instructions, so what reaches this event is always one
	// instruction with one amount, and the run's clause lives in
	// millUntilRunLocked between repetitions.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	// See mill.go.
	mill *millTail

	// --- RepEventCounter fields ---

	// CounterTarget / CounterName / CounterDelta are the counter
	// operation. Doubling Season doubles Delta; Hardened Scales adds
	// 1 when Name is "+1/+1"; Solemnity would set Canceled on +1/+1.
	//
	// CounterTarget is a CARD. A counter on a PLAYER sets CounterPlayer
	// below and leaves this at uuid.Nil; the two are never both set.
	CounterTarget uuid.UUID
	CounterName   string
	CounterDelta  int

	// CounterPlayer is set INSTEAD OF CounterTarget when the counters
	// go on a player (poison, energy, experience, rad). ADR 0056
	// Decision 5 keeps one event kind for both targets rather than
	// minting RepEventPlayerCounter, because the cards that replace
	// counters on players — Vorinclex, Lae'zel, Solemnity, Melira —
	// are the same cards that replace counters on permanents, and one
	// kind lets each keep one entry and one Watches value.
	//
	// Every counter replacement in the catalog that predates this
	// field opens its AppliesTo with g.LookupCardForEffect(ev.CounterTarget),
	// which fails for uuid.Nil, so none of them can apply to a player
	// event by accident. player_counters_test.go holds every
	// registered one to that, so a future card that forgets the check
	// fails CI rather than silently doubling somebody's poison.
	//
	// It is also the CR 616.1 affected player: see
	// affectedPlayerForEvent.
	CounterPlayer uuid.UUID

	// counterTail is the counter half's answer to lifeTail and
	// tokenTail: the rest of the effect that asked for the placement,
	// run with the delta the window settled on once the counters land.
	// Set by AddCounterByThenForEffect and read only by
	// runCounterTailLocked, which both the inline path and the CR 616
	// resume reach.
	//
	// #1282: a placement can PAUSE on a CR 616 ordering prompt (a
	// Doubling Season beside a Hardened Scales), and "amass, then the
	// Army deals damage equal to its power" cannot be written on the
	// next line.
	//
	// Unexported engine plumbing — the catalog never sets or reads it.
	// See counter_tail.go.
	counterTail *counterTail

	// CounterPlacer is who PUTS or GIVES the counters — the player
	// CR 120.3b and CR 120.3d name when damage from a source with
	// infect, wither or toxic becomes counters, and the proliferating
	// player for CR 701.34. uuid.Nil means UNKNOWN, not "nobody": a
	// "if YOU would put" replacement reads it first and falls back to
	// its own heuristic when it is Nil, which is weaker than printed
	// rather than stronger.
	//
	// It is NOT the same question as who controls the target. Vorinclex
	// keyed its halving clause on the target's controller for want of
	// this field and shipped a caveat saying so.
	//
	// A rules-driven counter REMOVAL (loyalty off from damage, the
	// stun-counter removal) bypasses this window entirely and carries
	// no placer: CR 614.1 has nothing to replace in a removal, and no
	// catalog card claims to.
	CounterPlacer uuid.UUID

	// CounterFromCombatDamage marks a placement that is the RESULT of
	// combat damage (CR 120.4c) rather than something an effect did.
	//
	// Doubling Season is why it exists: "if AN EFFECT would put one or
	// more counters" does not cover combat damage, which is a
	// turn-based action, so wither and infect COMBAT damage is not
	// doubled while a wither spell or a fight is. The passive "would be
	// put" cards (Hardened Scales, Winding Constrictor, Vizier of
	// Remedies) name no effect and are not gated on it.
	//
	// Set by the damage tail in ADR 0056's PR 2; nothing in the engine
	// sets it yet. The reader lands first on purpose — PR 2 is what
	// makes the wrong answer reachable, and a card that would start
	// doubling combat-damage counters the moment the tail branches is
	// not something to leave for the PR that branches it.
	CounterFromCombatDamage bool

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

	// SourceLKI is the damage source's characteristics AS THEY WERE
	// when the event was created — last-known information, CR 608.2h.
	// Nil when the source is unknown (a sandbox mark with no source,
	// an effect that names none).
	//
	// DamageSource alone cannot answer CR 702.16e, and not only
	// because of a pause. A SPELL source is never on the battlefield
	// at all, so a Lightning Bolt's redness has no lookup; and a
	// permanent that dealt damage and then died has DIFFERENT
	// characteristics in the graveyard from the ones it dealt the
	// damage with — a pumped, colour-shifted attacker is red on the
	// battlefield and colourless in the yard. Reading the new zone
	// would be the wrong object.
	//
	// Set from damageTail.sourceLKI by
	// damageThroughReplacementsLocked, the one body all six damage
	// entry points and the CR 616 resume go through, so there is
	// exactly one place it is filled in. See ADR 0072 §3.
	SourceLKI *Characteristic

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
	// Several things set it, and they fall into two families: paying a
	// cost (life, CR 118.3, payLifeAsCostLocked in life_tail.go; a
	// discard, CR 701.8a, discard.go, through zoneRoute.MustSettleNow;
	// a counter placement mid-ability, #1370,
	// AddCounterMustSettleNowForEffect in counter_tail.go) — CR 601.2h
	// pays a spell's costs as one indivisible step and CR 601.2 rewinds
	// the announcement if they cannot all be paid, so a CR 616 ordering
	// prompt in the middle would leave it half paid — and a mana
	// ability's own resolution (CR 605.3a, produce_mana.go), which has
	// no priority window at all for a prompt to occupy.
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

// MultiplyMana rewrites a RepEventProduceMana to produce n times as
// much of the same mana (CR 106.12b) — Mana Reflection's two,
// Nyxbloom Ancient's three.
//
// Each entry is repeated in place, so the multiplied production is
// still "that mana": a doubled {G} is {G}{G} and a doubled {C}{C} is
// {C}{C}{C}{C}. A factor of one is a no-op and a factor of zero or
// less produces nothing, which is the same outcome as ev.Cancel() and
// is the honest answer for a card that ever prints it.
//
// It is a method rather than a card-side loop because the whole family
// is one arithmetic operation on one field, and because a card that
// reached for `ev.ManaColors = ...` could silently change the COLOURS
// as well as the amount — which CR 106.12b does not license.
func (ev *ReplacementEvent) MultiplyMana(n int) {
	if ev == nil || ev.Kind != RepEventProduceMana || n == 1 {
		return
	}
	if n <= 0 {
		ev.ManaColors = nil
		return
	}
	out := make([]string, 0, len(ev.ManaColors)*n)
	for _, c := range ev.ManaColors {
		for k := 0; k < n; k++ {
			out = append(out, c)
		}
	}
	ev.ManaColors = out
}

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

	// EntryHandReveal, when non-nil, makes this a "you may reveal
	// <a card matching this> from your hand; if you don't,
	// <replacement>" effect — the ten reveal-lands, Choked Estuary
	// and its cycle. The apply-loop queues a
	// PendingChoiceEntryRevealFromHand pick and bails; revealing
	// means Replace NEVER runs, declining (or having nothing that
	// matches) means it does.
	//
	// EntryLifeCost's inversion with a CARD where the shockland has
	// a number, which is why it is not Optional either. A reveal is
	// not a cost and not a zone change (CR 701.20b): the named card
	// is still in hand afterwards. A player with no matching card is
	// not prompted and the replacement applies — the absence of the
	// question, not a refusal of it.
	//
	// Takes precedence over Optional, which is meaningless alongside
	// it. No printed card combines it with EntryLifeCost or
	// CopySelector. See entry_reveal.go and ADR 0013 §5z. Added with
	// the reveal-land cycle (#1198).
	EntryHandReveal *EntryHandReveal

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

	// Preemptive declares a RULES-LEVEL shield that applies before
	// any other applicable replacement, with no CR 616 ordering
	// prompt. Exactly one effect sets it: protection's damage
	// prevention (CR 702.16e, builtin_replacements.go).
	//
	// It exists because "which of these applies first?" has an
	// observable answer even when the event ends the same way.
	// Protection cancels the damage event outright, so in every
	// ordering it is the last thing to happen to that event — but a
	// CHARGED prevention shield ("prevent the next 4 damage",
	// effects.PreventNextDamage) ordered first would spend a charge
	// absorbing damage that was never going to be dealt. #420 is
	// that bug; this flag is the fix.
	//
	// It is a DECLARED SIMPLIFICATION of CR 616.1, which gives the
	// affected object's controller the ordering choice. See ADR 0072
	// §4 for why the prompt is taken away and what it costs
	// (Phytohydra).
	Preemptive bool

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
// object dropped. `ReplacementEffectID` packs the battlefield index
// and the slot (encodeCatalogReplacementID), so it says "the Doubling
// Season in slot 3 of the battlefield"; this says "Doubling Season's
// counter-doubling replacement, controlled by Ian" — the same for
// every copy on the board. That is the whole point: two Rhox
// Faithmenders are two objects contributing ONE effect, and #792 is
// about not asking which of two identical modifications happened
// first.
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

// encodeCatalogReplacementID mints the per-instance ID for slot
// `slot` of the catalog card at battlefield index `cardIdx`: the two
// coordinates packed into one ReplacementEffectID with
// MaxCatalogReplacementSlots as the stride, offset by one so no live
// ID is zero (the zero value means "no ID").
//
// ok is false for anything the scheme cannot represent — a negative
// index, a slot past the budget, or a card index far enough along the
// battlefield to run into the self-replacement range above. A Spec
// over the budget is refused by effects.Register at boot, so a false
// here means a test stub or a battlefield nobody could reach; the
// gather pass skips the entry rather than mint an ID that names a
// different card's effect.
//
// decodeCatalogReplacementID is the exact inverse. The pair is the
// only place the packing is written (#801) — it used to be a literal
// in the gather pass and a second, hand-rolled literal in
// ReplacementOptionMetaForEffect, two sites that had to change
// together and no check that they did.
func encodeCatalogReplacementID(cardIdx, slot int) (ReplacementEffectID, bool) {
	if cardIdx < 0 || slot < 0 || slot >= MaxCatalogReplacementSlots {
		return 0, false
	}
	id := ReplacementEffectID(cardIdx)*MaxCatalogReplacementSlots + ReplacementEffectID(slot) + 1
	if id >= selfReplacementIDBase || id < 1 {
		return 0, false
	}
	return id, true
}

// decodeCatalogReplacementID unpacks a catalog ReplacementEffectID
// back into the battlefield index and slot it was minted from.
//
// ok is false for an ID outside the catalog range — zero, or one
// belonging to a self-replacement, a turn-scoped effect, a test
// injection or a built-in. Callers that handle those ranges check
// them first; this is the defensive floor for a stale or malformed ID
// off the wire, which decodes to nothing rather than to whatever card
// happens to sit at that battlefield index.
//
// The returned cardIdx is NOT bounds-checked against the battlefield:
// the ID space is far larger than any battlefield, so only the caller
// holding the slice can say whether the index is live.
func decodeCatalogReplacementID(id ReplacementEffectID) (cardIdx, slot int, ok bool) {
	if id < 1 || id >= selfReplacementIDBase {
		return 0, 0, false
	}
	raw := id - 1
	return int(raw / MaxCatalogReplacementSlots), int(raw % MaxCatalogReplacementSlots), true
}

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
		// A PREEMPTIVE effect applies before anything else and
		// without a prompt, however many others also apply — the
		// whole point is that nothing else gets to charge itself
		// first (ReplacementEffect.Preemptive; protection, CR
		// 702.16e). Built-ins are gathered first, so scanning from
		// the front finds it immediately on the ordinary board.
		if i := firstPreemptive(applicable); i >= 0 {
			g.applyFirstGatheredLocked(ev, applicable[i:i+1])
			continue
		}
		if len(applicable) > 1 {
			// CR 616: affected player picks order. Queue a prompt
			// and stash the resume frame; caller returns without
			// applying — with four exceptions, all of which apply
			// the gathered order inline instead, one effect per pass
			// round this loop (CR 616.1f; applyFirstGatheredLocked).
			//
			// #710: every applicable effect is a PURE CANCEL, so
			// every ordering produces the same event. Two "skip your
			// draw step" enchantments under one controller are one
			// skipped draw step, and asking which of them skipped it
			// is a prompt with a single answer.
			if allPureCancels(applicable) {
				g.applyFirstGatheredLocked(ev, applicable)
				continue
			}
			// #792: every applicable effect is the SAME declared
			// effect on two or more objects — two Doubling Seasons,
			// two Rhox Faithmenders. The player is being asked to
			// order N copies of one modification, so again there is
			// only one answer.
			if sameModification(applicable) {
				g.applyFirstGatheredLocked(ev, applicable)
				continue
			}
			// Nobody is left to answer the prompt: the gathered
			// order stands (an eliminated player's prompt would
			// block the table forever; S31 fuzzer finding).
			//
			// #847: through skipQuestionsLocked, exactly as the
			// mustSettleNow branch below does. An effect that asks its
			// own question cannot ride an order nobody chose either —
			// firing a CR 614.10 "may" or a copy selector here would
			// answer it blind, in the direction that favours it, on
			// an event whose affected player is no longer at the
			// table. Skipped un-applied is the weaker branch, which is
			// the posture every other un-prompted path takes.
			if chooser := affectedPlayerForEvent(ev, applicable, g); g.chooserGoneLocked(chooser) {
				g.applyFirstGatheredLocked(ev, g.skipQuestionsLocked(ev, applicable))
				continue
			}
			// #793: the event cannot pause — a life payment is a COST
			// (CR 118.3) and CR 601.2h pays a spell's costs as one
			// indivisible step. The gathered order stands, for the same
			// reason it does above: a prompt nobody can answer here is
			// a spell stuck on the stack with a half-paid cost.
			if ev.mustSettleNow {
				g.applyFirstGatheredLocked(ev, g.skipQuestionsLocked(ev, applicable))
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
		if chosen.effect.EntryHandReveal != nil {
			// "As this enters, you may reveal an Island or Swamp
			// card from your hand." The same shape one line up with
			// a card where the shockland has a number: the prompt
			// and its three apply-it-inline cases live in
			// entry_reveal.go, a queued prompt bails and an inline
			// apply falls through to the next iteration (#1198,
			// ADR 0013 §5z).
			if g.offerEntryHandRevealLocked(ev, chosen) {
				return ev, errReplacementPending
			}
			continue
		}
		if chosen.effect.Optional {
			// CR 614.10 "may" — owner decides each time. Queue a
			// yes/no prompt; the resume path either fires Replace
			// (yes) or marks applied and skips (no). The two cases
			// that decline it inline instead — a chooser who has left
			// the game, an event with nothing to resume it (#359) —
			// live in the helper with the prompt, because
			// ResolveReplacementOrder's chosen-order loop needs the
			// same three answers and used to have none of them
			// (#847). A queued prompt bails; a decline falls through
			// to the next iteration.
			if g.offerOptionalReplacementLocked(ev, chosen) {
				return ev, errReplacementPending
			}
			continue
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

// firstPreemptive is the index of the first gathered replacement
// that declared itself preemptive, or -1. See
// ReplacementEffect.Preemptive.
func firstPreemptive(applicable []activeReplacement) int {
	for i, a := range applicable {
		if a.effect.Preemptive {
			return i
		}
	}
	return -1
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
// "may", a shockland's pay-life, a reveal-land's pick from hand, a
// copy selector.
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
	return e.Optional || e.EntryLifeCost > 0 || e.EntryHandReveal != nil || e.CopySelector != nil
}

// applyFirstGatheredLocked fires the FIRST gathered replacement and
// marks it applied. The un-prompted sibling of ResolveReplacementOrder's
// chosen-order loop, used when the CR 616 prompt is skipped — because
// nobody could answer it (an eliminated chooser), because every
// ordering gives the same answer (allPureCancels, sameModification), or
// because the event cannot pause (mustSettleNow).
//
// Only the first, and the caller goes round the apply-loop again, which
// re-gathers. CR 616.1f: "Once the chosen effect has been applied, this
// process is repeated (taking into account only replacement or
// prevention effects that would now be applicable)". This used to fire
// the whole gathered list back to back, so an effect whose AppliesTo
// the first one had just made false — a "lose life greater than 3"
// gate after a halving, say — fired anyway on a stale gather (#808).
// Going round again also lets an effect the first one ENABLED join in
// at its proper place instead of after the rest of the list. The
// orderings the skipped prompt would have offered are still not
// offered — the gather order stands — which is the declared
// simplification those branches already carry.
//
// An empty list is a no-op: skipQuestionsLocked can drop every entry,
// and the next gather then finds them marked.
//
// Caller must hold g.mu, and must have allocated the once-per-event
// map entry for ev.ID.
func (g *Game) applyFirstGatheredLocked(ev *ReplacementEvent, applicable []activeReplacement) {
	if len(applicable) == 0 || ev.Canceled {
		return
	}
	chosen := applicable[0]
	g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
	if chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
}

// skipQuestionsLocked drops the gathered replacements that would ask
// their controller a question — a CR 614.10 "may", a shockland's
// pay-life, a reveal-land's pick from hand, a copy selector —
// marking each applied so the apply-loop
// does not gather it again, and returns the rest in gather order.
//
// Two windows use it, and both are windows in which the question
// could not be put to anybody anyway: an event that cannot pause
// (mustSettleNow), and an ordering window whose affected player has
// left the game (#847). Everywhere else the question is asked.
// Skipping is the weaker branch of a "may" and the un-copied branch of
// a selector, which is what "never stronger than printed" means here.
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

// stillAppliesLocked reports whether a gathered replacement would
// still be gathered for ev as the effects applied so far have left it:
// it has not fired for this event, and its Watches filter and AppliesTo
// predicate still pass.
//
// CR 616.1f re-checks applicability after every applied effect ("taking
// into account only replacement or prevention effects that would now be
// applicable"). The CR 616 ordering answer fires the whole chosen order
// in one go, so it asks this before each effect rather than trusting
// the gather the prompt was built from (#808).
//
// It tests the entry the frame is holding rather than re-gathering and
// looking the ID up, because a catalog effect's ID is its battlefield
// index: a permanent leaving between the prompt and the answer shifts
// every ID after it, and a lookup would then answer for a different
// effect.
//
// Caller must hold g.mu.
func (g *Game) stillAppliesLocked(ev *ReplacementEvent, a activeReplacement) bool {
	if ev == nil || g.replacementsAppliedThisEvent[ev.ID][a.id] {
		return false
	}
	if !eventKindMatches(a.effect.Watches, ev.Kind) {
		return false
	}
	return a.effect.AppliesTo == nil || a.effect.AppliesTo(ev, g, a.source)
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
	// IDs are minted by encodeCatalogReplacementID, which packs the
	// battlefield index and the slot into the bottom of the ID space.
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
				id, ok := encodeCatalogReplacementID(cardIdx, repIdx)
				if !ok {
					// Unrepresentable — a slot past the budget, which
					// effects.Register refuses at boot. Skipping is
					// the only safe answer: a minted-anyway ID would
					// name another card's effect.
					continue
				}
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
	//
	// And not for a card entering FACE DOWN (#1322). CR 614.12 decides
	// which replacements apply from the permanent "as it would exist on
	// the battlefield", and a face-down permanent is a 2/2 with no
	// abilities (CR 708.2a). A face-down CAST already reads "" here —
	// the card is face down on the stack — but a manifest takes a card
	// that is face UP in its library, so without this a manifested
	// shockland asked its controller for 2 life, on a card the rest of
	// the table was not allowed to see. That question could not be put
	// while the put batch was unresumable, so it took the un-asked
	// branch and the manifest entered TAPPED instead — wrong the other
	// way, and silently.
	if CatalogReplacements != nil && ev.CardID != uuid.Nil && !g.Battlefield.Contains(ev.CardID) &&
		!(ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && ev.FaceDown != FaceDownNone) {
		if entering, ok := g.LookupCardForEffect(ev.CardID); ok {
			key := CatalogKey(entering)
			reps := CatalogReplacements(key)
			for repIdx := range reps {
				if repIdx >= MaxCatalogReplacementSlots {
					// The same budget the packed catalog IDs are
					// bounded by, for the same reason: slot
					// MaxCatalogReplacementSlots would be the first
					// turn-scoped effect's ID.
					break
				}
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
	// Catalog — the packed (battlefield index, slot) range, unpacked
	// by the same pair the gather pass mints with.
	if CatalogReplacements == nil {
		return "", uuid.UUID{}
	}
	cardIdx, repIdx, ok := decodeCatalogReplacementID(id)
	if !ok {
		// Below the catalog range (zero) or in the self-replacement
		// range, which names a card that is not on the battlefield
		// and so has no slot to look up here.
		return "", uuid.UUID{}
	}
	if cardIdx >= len(g.Battlefield.Cards) {
		return "", uuid.UUID{}
	}
	card := &g.Battlefield.Cards[cardIdx]
	reps := CatalogReplacements(CatalogAbilityKey(*card))
	if repIdx >= len(reps) {
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
	case RepEventDiscard:
		// CR 701.8a. A discard is a move out of the hand, but what a
		// discard replacement watches for is the DISCARD — Library of
		// Leng, madness, "if you would discard a card, exile it
		// instead" — so it keys on the discard event, not the zone
		// move. An effect that wants both declares both.
		want = EventDiscardCard
	case RepEventCreateTokens:
		want = EventTokenCreated
	case RepEventKeywordAction:
		// CR 701.22 / 701.25 / 701.34. One watch key for all three
		// counted keyword actions, because the action is a field on
		// the event and the card-side helper narrows on it — see
		// EventKeywordAction, which is a replacement-watch sentinel
		// like EventStepTransition rather than a logged event.
		want = EventKeywordAction
	case RepEventMill:
		// CR 701.13a. EventMill is the post-event twin and it fires per
		// CARD, after the move, while this window is one per
		// INSTRUCTION and opens before any card has left the library.
		// Reusing the key rather than minting a sentinel is what
		// RepEventCreateTokens and RepEventDiscard already do with
		// EventTokenCreated and EventDiscardCard, whose twins fire
		// afterwards too: an EventKind means "the mill" in a
		// ReplacementEffect.Watches and "a card was milled" in a
		// TriggeredAbility.Watches, and no code reads one as the other.
		// The sentinel spelling is for a family with no twin at all
		// (EventStepTransition, EventKeywordAction).
		want = EventMill
	case RepEventProduceMana:
		// CR 106.12b. EventManaAdded is the post-event twin and it
		// fires per MANA, after the token is in the pool, while this
		// window is one per production and opens before any of it is
		// there. Reusing the key rather than minting a sentinel is what
		// RepEventMill already does with EventMill, and for the same
		// reason: an EventKind means "the production" in a
		// ReplacementEffect.Watches and "a mana was added" in a
		// TriggeredAbility.Watches, and no code reads one as the other.
		want = EventManaAdded
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
