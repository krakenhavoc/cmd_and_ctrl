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
//     route is a discard (CR 701.9, #650)
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
//	"discard"      — RepEventDiscard — CardID, DiscardPlayer, DiscardCause, NewZone, NewZoneOwner (CR 701.9)
//	"counter"      — RepEventCounter — CounterTarget OR CounterPlayer,
//	                 CounterName, CounterDelta, CounterPlacer,
//	                 CounterFromCombatDamage
//	"life"         — RepEventLife    — LifePlayer, LifeDelta
//	"damage"       — RepEventDamage  — DamageSource, DamageTarget, DamageAmount, IsCombatDamage
//	"create_tokens"— RepEventCreateTokens — TokenController, TokenGroups, TokenAttacking (CR 701.7b)
//	"keyword_action"— RepEventKeywordAction — KeywordAction, KeywordActionCount (CR 701.22 / 701.25 / 701.34)
//	"mill"         — RepEventMill    — MillPlayer, MillCount (CR 701.17a)
//	"step"         — RepEventStepTransition — StepTransitionStep, StepTransitionSeat
type ReplacementEventKind string

const (
	RepEventDraw    ReplacementEventKind = "draw"
	RepEventMove    ReplacementEventKind = "move"
	RepEventCounter ReplacementEventKind = "counter"
	RepEventLife    ReplacementEventKind = "life"
	RepEventDamage  ReplacementEventKind = "damage"

	// RepEventDiscard is a discard (CR 701.9a) — the one exit whose
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

	// RepEventMill is one "mill N cards" INSTRUCTION (CR 701.17a),
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
	// Opened only for a real mill: a graveyard destination (CR 701.17a
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
	// It is the one kind that can NEVER pause. CR 605.3b makes
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
//	selfReplacementIDBase   .. scopedReplacementIDBase  a card replacing its own entry, by slot,
//	                                                    then riot and unleash by instance
//	                                                    (entryKeywordReplacementID, riot.go)
//	scopedReplacementIDBase .. testReplacementIDBase    a ScopedEffect's replacement mod, by Seq × mod
//	testReplacementIDBase   .. builtinReplacementIDBase  test-injected, by index
//	builtinReplacementIDBase ..                     built-ins (commander zone), by index
//
// They were three `const` declarations inside the two functions that
// read them, declared twice over; #801 moved them here so the scheme
// is written down once.
const (
	selfReplacementIDBase    ReplacementEffectID = 1 << 45
	scopedReplacementIDBase  ReplacementEffectID = 1 << 50
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

	// drawInstead is the body of a draw replaced by an EFFECT rather
	// than by a number: dredge's "mill N cards and return this card",
	// Underrealm Lich's "look at the top three cards, put one into your
	// hand and the rest into your graveyard", Forbidden Crypt's "return
	// a card from your graveyard to your hand". Set when a replacement
	// declaring ReplacementEffect.DrawInstead is applied, which also
	// cancels the draw. See draw_instead.go.
	//
	// It is a registered KEY plus the source card, not a function: the
	// body is looked up when the window settles, so the event holds
	// plain data and a snapshot taken across a prompt carries nothing a
	// clone could share.
	drawInstead drawInsteadRun

	// drawTail is how many individual draws of the SAME instruction are
	// still owed once this one is finished — "draw three" paused on its
	// first card owes two. Unexported engine plumbing; the catalog
	// neither sets nor reads it.
	drawTail int

	// drawThen is the rest of the card after "draw …, then": plain
	// data (a registered body's key and scalars), run once the last
	// draw of the instruction is done. See DrawNThenForEffect.
	drawThen DrawThen

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

	// ShuffleDestinationLibrary is a replacement's declaration that its
	// redirected NewZone == ZoneLibrary is a SHUFFLE-in, not merely a
	// placement — "shuffle it into its owner's library instead"
	// (Blightsteel Colossus, the Eldrazi titans' graveyard clause),
	// as opposed to Library of Leng's "put it on top" or a tuck's
	// unshuffled bottom/depth placement.
	//
	// Set by the SAME `Replace` that rewrites NewZone, never derived:
	// nothing about a library destination says whether the printed
	// text shuffled or placed, and guessing would either shuffle a
	// Library of Leng discard (wrong) or leave a Blightsteel on top
	// (the S17-era caveat this field closes).
	//
	// Consumed once, after the move has actually landed in a library —
	// never before, and never on a destination a further replacement
	// or a missing player redirected elsewhere — by
	// executeZoneRouteLocked and executeBattlefieldLeaveLocked, the
	// two functions that perform every replaced move's physical
	// landing. A card that never reaches the library it was aimed at
	// (canceled, or resolved to exile for a departed owner) shuffles
	// nothing. See ADR 0013 §5ah.
	ShuffleDestinationLibrary bool

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
	// inside its Replace, because CR 701.19c leaves an ignored shield
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

	// EntersDevoured is the number of creatures devoured as this
	// permanent enters (CR 702.82a), stamped by a devour clause's
	// EntryCardChoice.Devour and copied onto Card.Devoured by
	// every battlefield landing. Rides the event for EntersTapped's
	// reason: the permanent is not on the battlefield while the
	// sacrifice is chosen.
	EntersDevoured int

	// EntersWithHaste is riot's "if you don't, it gains haste" (CR
	// 702.136a, ADR 0109 §10): set by the haste answer to an entry_riot
	// question and read by every battlefield landing, which gives the
	// permanent an indefinite haste record pinned to it once its entry
	// has been stamped and before EventETB (grantRiotHasteLocked). Rides
	// the event for EntersTapped's reason. Only meaningful when NewZone
	// == ZoneBattlefield. Haste is redundant (CR 702.10d), so two riots
	// answered "haste" are one record.
	EntersWithHaste bool

	// lookAhead caches the CR 614.12 look-ahead for this entry
	// (entry_lookahead.go), so the gather, which runs once per pass of
	// the apply-loop, does not run a dry layer pass every time. It is
	// keyed on what can change the answer mid-window.
	lookAhead *entryLookAheadCache

	// EntersUnlocked is CR 709.5d's designation for a Room spell that
	// resolves (ADR 0103): the door of the half that was cast, seeded
	// from the resolving stack item before the CR 614 pipeline and
	// given to the permanent by every battlefield landing before
	// EventETB, so the unlocked door's statics apply as it enters. Zero
	// for every other entry — a Room put onto the battlefield without
	// being cast enters with both doors locked.
	EntersUnlocked DoorMask

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

	// copyUntilEndOfTurn marks EntersAsCopyOf as a copy WITH A
	// DURATION — "as this enters, you may have it become a copy of …
	// until end of turn" (Cursed Mirror, #1593). The entry path lands
	// it like any entry copy, and settleTimedEntryCopyLocked re-files
	// it as a duration copy once the permanent has its entry stamp.
	copyUntilEndOfTurn bool

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

	// landPlayer is the player whose land drop a landPlay entry
	// spends, captured before the window opens. ev.Actor is the
	// would-be CONTROLLER, and an entry-controller effect (ADR 0102)
	// rewrites it; the land drop belongs to whoever PLAYED the land
	// whatever it enters under, so the tally reads this and not Actor.
	// uuid.Nil falls back to Actor, which is every entry built before
	// the field existed.
	landPlayer uuid.UUID

	// entryControllerSet and entryControllerPrior are an
	// entry-controller effect's bookkeeping (ADR 0102 decision 2): the
	// first time one re-stamps the entering card's controller, what the
	// card carried is remembered so a cancelled entry can put it back.
	// See setEntryControllerLocked.
	entryControllerSet   bool
	entryControllerPrior uuid.UUID

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

	// commanderAnswer is the CR 903.9 answer the commander's owner gave
	// before this move was made (#1397, cost_commander_choice.go),
	// copied off zoneRoute.commanderAnswer. Unasked leaves the built-in
	// as it always was; accept makes it mandatory for this event, so a
	// cost move that cannot pause applies it instead of skipping it as
	// a question; decline keeps it from being gathered at all.
	commanderAnswer commanderZoneAnswer

	// --- RepEventDiscard fields ---
	//
	// A discard also fills in the RepEventMove fields above: CardID is
	// the discarded card, OldZone is the hand it is leaving, and
	// NewZone / NewZoneOwner are the destination a replacement may
	// rewrite.

	// DiscardPlayer is the player discarding the card — its owner,
	// because every hand in this engine holds only its owner's cards
	// (CR 701.9a moves the card to that player's graveyard).
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
	// as many as possible (CR 701.17b) and the clamp happens in
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

	// CounterFromCost marks a placement that PAYS A COST rather than
	// one an effect makes: a blight paid as an additional cost to cast
	// (#1703, CR 701.68a at CR 601.2h).
	//
	// It is CounterFromCombatDamage's sibling, for the same card and
	// the same rule. CR 614.16 says a replacement worded "if an effect
	// would put one or more counters" applies to the effect of a
	// resolving spell or ability. A cost payment is neither, so
	// Doubling Season does not double a blight. A replacement that
	// names no effect (Vorinclex, Winding Constrictor, Vizier of
	// Remedies) replaces the event itself and is not gated on this.
	// That is the Devoted Druid + Vizier ruling, and it is why a
	// cost's counters open this window at all.
	CounterFromCost bool

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

	// DamageInstance is the instance of damage this event belongs to
	// (ADR 0108 PR 0, damage_instance.go): every event one damage
	// instruction or one combat damage step opens carries the same one,
	// so a reader can tell "at the same time" from "one after the other
	// inside one resolution" (CR 615.8, 608.2c). Stamped by the engine's
	// damage entry points as the event is opened; a replacement reads
	// it and never writes it. It rides the event, so a CR 616 prompt
	// resumes with it. Zero on an event built by hand, which readers
	// treat as the ADR 0107 event batch. Transient, never captured.
	DamageInstance DamageInstance

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
	// discard, CR 701.9a, discard.go, through zoneRoute.MustSettleNow;
	// a counter placement mid-ability, #1370,
	// AddCounterMustSettleNowForEffect in counter_tail.go) — CR 601.2h
	// pays a spell's costs as one indivisible step and CR 601.2 rewinds
	// the announcement if they cannot all be paid, so a CR 616 ordering
	// prompt in the middle would leave it half paid — and a mana
	// ability's own resolution (CR 605.3b, produce_mana.go), which has
	// no priority window at all for a prompt to occupy.
	//
	// What it costs the affected player is the CR 616 ordering choice
	// and any "may" on the event: the apply-loop applies the
	// gathered order inline (the escape an eliminated chooser has
	// always taken) and skips anything that would ask a question
	// un-applied. Weaker than printed on the "may", arbitrary but
	// deterministic on the ordering, and never a wedged table.
	//
	// The one "may" a cost does NOT lose is CR 903.9's: the owner of a
	// commander a cost moves is asked BEFORE the payment begins, and
	// the answer rides in as commanderAnswer (#1397,
	// cost_commander_choice.go).
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
	// ActiveWhen is ADR 0071's designation gate: the replacement exists
	// only while the gate is satisfied by the permanent it is printed
	// on. Its zero value is "no gate", which is every replacement in the
	// catalog but a Room door's (ADR 0103 — Torture Pit's "if a source
	// you control would deal noncombat damage …" exists only while its
	// door is unlocked). Read by the battlefield gather and the entering
	// card's self-replacement gather, which skip an inactive slot
	// rather than filter the list, so a slot's index — which the
	// replacement's ID packs — never moves.
	ActiveWhen Designation

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

	// Optional flags a "may" replacement — the owner
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
	// A player who can't legally pay (CR 119.4 — life total below
	// the cost) is not prompted; the replacement applies. Takes
	// precedence over Optional, which is meaningless alongside it.
	// See entry_choice.go. Added with the shockland cycle.
	EntryLifeCost int

	// EntryCardChoice, when non-nil, makes this a "<reveal / discard /
	// sacrifice> <cards matching this>; if you don't, <replacement>"
	// effect — the ten reveal-lands (#1198), Mox Diamond and the seven
	// "sacrifice … instead" lands (#1744, ADR 0098). The apply-loop
	// queues a card-set pick and bails; naming cards (and, for a
	// discard or a sacrifice, their really leaving) means Replace
	// NEVER runs, declining — or having nothing that matches — means
	// it does.
	//
	// EntryLifeCost's inversion with a CARD where the shockland has
	// a number, which is why it is not Optional either, and why
	// declineIsReplace lists it: in a window that cannot ask, the
	// weaker branch is to RUN Replace.
	//
	// Takes precedence over Optional, which is meaningless alongside
	// it. No printed card combines it with EntryLifeCost or
	// CopySelector. See entry_card_choice.go, ADR 0013 §5z and
	// ADR 0098.
	EntryCardChoice *EntryCardChoice

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

	// EntryController, when non-nil, makes this a "this permanent
	// enters under the control of an opponent of your choice" effect
	// (CR 614.1d, CR 614.12) — Captive Audience, Pendant of
	// Prosperity, Abby, Merciless Soldier, Xantcha, Sleeper Agent. The
	// apply-loop settles it inline or queues a
	// PendingChoiceEntryController and bails; the answer becomes the
	// would-be controller the permanent lands under. Replace is never
	// called. Mandatory: an entry that cannot pause takes the default
	// rather than skipping it. See entry_controller.go and ADR 0102.
	EntryController *EntryControllerChoice

	// ChangesEntryController declares CR 616.1b's tier: this effect
	// would modify under whose control an object enters, so it is
	// applied before every other effect in the window, and when
	// several such effects apply only they are ordered
	// (entryControlTier). Declared by the constructor, never inferred
	// from Replace. Every EntryController effect sets it; a later
	// "enters under your control instead" (Gather Specimens) would set
	// it too.
	ChangesEntryController bool

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

	// Prevention declares a PREVENTION effect (CR 615.1a: an effect
	// that uses the word "prevent"): one that watches a damage event
	// and prevents some or all of it. Protection's built-in (CR
	// 702.16e), Fog's and Mending Hands' scoped shields and every
	// catalog "prevent all damage that would be dealt to …" set it.
	//
	// It is what CR 615.12 reads. While a damage event can't be
	// prevented (damageUnpreventableLocked, unpreventable_damage.go),
	// a prevention effect is still applied to it, once (CR 615.12a),
	// and prevents nothing: its Replace is not run, so the damage and
	// a charged shield's charge are both left as they were. A
	// replacement that is NOT a prevention effect — a doubler, a
	// redirection, a shield counter's "remove a counter instead" — is
	// untouched by "can't be prevented", which is why the flag is
	// declared rather than inferred from what Replace does.
	//
	// TestDamageReplacementsDeclareWhetherTheyPrevent
	// (cards/effects) fails a damage replacement whose Replace cancels
	// or reduces the damage without declaring this or RedirectsDamage.
	Prevention bool

	// RedirectsDamage declares a REDIRECTION: a replacement that has
	// damage "dealt instead" to another permanent or player (CR
	// 614.1a). It is not a prevention effect (CR 615.12 does not stop
	// it), but "that damage can't be … dealt instead to another
	// permanent or player" (Lava Burst, Whippoorwill) does: while an
	// event can't be redirected (damageCantBeRedirectedLocked), a
	// redirection is applied without running its Replace, exactly as a
	// prevention effect is under CR 615.12.
	//
	// ADR 0108 §9: a redirection's Replace changes the recipient only
	// through RedirectDamageEventForEffect (redirect_damage.go), which
	// rewrites the target and the damage tail together and does nothing
	// when CR 614.9 says so; TestDamageReplacementsDeclareWhetherTheyPrevent
	// fails a Replace that writes DamageTarget by hand, and one that calls
	// the primitive without declaring this. A redirection has no shared
	// identity (catalogReplacementIdentity): two copies write two
	// different objects into the event, so they are ordered, not merged.
	RedirectsDamage bool

	// Then is a prevention static's CR 615.5 additional effect (ADR 0108
	// §8, #1906): a registered body, run after the prevention with what
	// it prevented — Nine Lives's incarnation counter, Immortal Coil's
	// "exile a card from your graveyard for each 1 damage prevented this
	// way", a Phantom's "remove a +1/+1 counter". The zero BodyRef is
	// none. Catalog data, never a closure, and never captured.
	//
	// It requires Prevention (effects.Register refuses it otherwise),
	// and its Replace may only prevent: the apply loop measures the
	// event before and after Replace and owes the difference
	// (prevention_then.go), and CR 615.12 owes it with nothing prevented
	// when the damage can't be prevented. TestPreventionStaticsOnlyPrevent
	// (cards/effects) fails a Prevention static whose Replace does
	// anything but change the event.
	Then BodyRef

	// ThenPer is the unit Then runs once per, within one damage
	// instance (ADR 0108 §8 decision 2), and is required with Then.
	// The printed subject picks it: "if damage would be dealt to X" is
	// one application per RECIPIENT (the Phantoms: blocked by three
	// creatures, one +1/+1 counter is removed), "if a source would deal
	// damage to X" one per SOURCE (Nine Lives: "if more than one source
	// deals damage to you at once … put that many incarnation
	// counters").
	ThenPer PreventionUnit

	// PromptQuestion is the text rendered in the yes/no Optional
	// prompt. Short — fits in a modal header. Defaults to Label
	// when empty.
	PromptQuestion string

	// Label is the CR 616 prompt's header copy
	// ("Doubling Season: double counters"). Kept server-side so
	// the wire carries it; no localisation yet.
	Label string

	// entryKeyword names the keyword an engine-derived entry
	// replacement stands for — KeywordRiot or KeywordUnleash (ADR 0109
	// §10, riot.go). Empty for every catalog, scoped and built-in
	// replacement. Unexported: the catalog grants the KEYWORD, and the
	// gather derives the replacement from the entry look-ahead.
	entryKeyword string

	// DrawInstead, on a RepEventDraw replacement, replaces the draw with
	// an effect that may need to ask something (draw_instead.go). The
	// effect is applied like any other, but instead of a Replace it
	// cancels the draw and, once the window has SETTLED, runs the
	// registered body (RegisterDrawInstead), which may queue prompts and
	// calls `done` when it has finished so the rest of a multi-card draw
	// waits behind it (CR 121.6b).
	//
	// Leave Replace nil: the engine cancels the draw itself. The zero
	// value is "not a draw-substituting effect".
	DrawInstead DrawInsteadRef

	// FromGraveyard makes this a replacement its source applies from
	// the DRAWING PLAYER's graveyard rather than from the battlefield —
	// dredge (CR 702.52a), the one ability a card has in that zone. The
	// battlefield walk skips such an effect, and the gather finds it in
	// the graveyard instead. Only a draw event consults the graveyard
	// (draw_instead.go).
	FromGraveyard bool

	// commanderZone marks the CR 903.9 built-in
	// (commanderZoneReplacement) so the gather can honour an answer
	// its owner gave BEFORE the move (ReplacementEvent.commanderAnswer,
	// #1397). Unexported: no catalog effect is the commander rule.
	commanderZone bool
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
// with anything". Built-in, scoped (ADR 0041 P8) and test replacements take
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
// note in docs/adding-cards.md tells card authors to flag it rather than
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

// catalogReplacementIdentity is a catalog replacement's identity — or the
// zero identity for a REDIRECTION (ADR 0108 §9). "All damage that would be
// dealt to you is dealt to enchanted creature instead" writes its own
// object's creature into the event, so two Pariahs on two creatures are
// two different modifications, and the affected player orders them
// (CR 616.1). The zero identity never matches, so sameModification never
// collapses them.
func catalogReplacementIdentity(eff ReplacementEffect, key string, slot int, controller uuid.UUID) replacementIdentity {
	if eff.RedirectsDamage {
		return replacementIdentity{}
	}
	return replacementIdentity{card: key, slot: slot, controller: controller}
}

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
// belonging to a self-replacement, a scoped effect, a test
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
		// CR 615.12: a prevention effect applied to damage that can't
		// be prevented prevents nothing, and is applied only once
		// (CR 615.12a). Settled here, before any prompt is built, so a
		// shield that will do nothing is never an ordering question.
		applicable = g.settleUnpreventableLocked(ev, applicable)
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
		// CR 616.1b (ADR 0102): an effect that would change under whose
		// control the object enters is applied before any other, so
		// only those are candidates on this pass. The rest are gathered
		// again afterwards, against the new controller (CR 616.1f).
		applicable = entryControlTier(applicable)
		// CR 702.136b (#1556): instances of one entry keyword — a
		// printed riot and Rhythm of the Wild's — each work
		// separately, and which is asked first changes nothing. The
		// first is asked alone, and the rest are gathered again
		// afterwards, as #792 does for two copies of one effect.
		if len(applicable) > 1 && sameEntryKeyword(applicable) {
			applicable = applicable[:1]
		}
		// Dredge cards in one graveyard (draw_instead.go): each is its
		// own "may", so ordering them is not a question. Asking each in
		// turn lets the player pick any one of them or none, and the
		// first "yes" ends the draw.
		if len(applicable) > 1 && allGraveyardOptions(applicable) {
			applicable = applicable[:1]
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
			// firing a "may" or a copy selector here would
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
		if chosen.effect.EntryController != nil {
			// "Enters under the control of an opponent of your choice"
			// (ADR 0102). Mandatory, so it is asked BEFORE the
			// cannot-pause skip below: an entry that cannot pause takes
			// the default opponent rather than skipping the effect.
			// entry_controller.go settles the inline cases; a queued
			// prompt bails.
			if g.offerEntryControllerLocked(ev, chosen) {
				return ev, errReplacementPending
			}
			continue
		}
		if ev.mustSettleNow && asksItsOwnQuestion(chosen.effect) {
			// #793, the single-effect half of the branch above: an
			// effect that would ask its controller a question cannot
			// fire on an event that cannot pause, and firing it blind
			// would answer the question for them in the direction that
			// favours it. Settled on its weaker branch — skipped for a
			// "may", its decline run for a shockland, a reveal-land or
			// a Mox Diamond (ADR 0098 Decision 5).
			g.skipOwnQuestionLocked(ev, chosen)
			continue
		}
		// An effect with a question inside it — a copy selector, a
		// shockland's life, an entry card choice, a "may". The prompt
		// and each shape's apply-it-inline cases live with the shape;
		// a queued prompt bails, anything else has already marked the
		// effect applied and falls through to the next iteration.
		if pending, handled := g.offerOwnQuestionLocked(ev, chosen); handled {
			if pending {
				return ev, errReplacementPending
			}
			continue
		}
		// Mandatory: fire Replace inline and iterate.
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		g.runReplaceLocked(ev, chosen)
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
		if !a.effect.PureCancel || asksItsOwnQuestion(a.effect) || a.effect.EntryController != nil {
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
		if a.identity != want || asksItsOwnQuestion(a.effect) || a.effect.EntryController != nil {
			return false
		}
	}
	return true
}

// asksItsOwnQuestion reports whether firing this effect puts a
// question to a player before it changes anything — a
// "may" replacement, a shockland's pay-life, a reveal-land's pick from hand, a
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
	return e.Optional || e.EntryLifeCost > 0 || e.EntryCardChoice != nil || e.CopySelector != nil ||
		e.entryKeyword == KeywordRiot
}

// declineIsReplace reports that this effect's Replace IS the "you
// didn't" branch of its question — the shockland's enters-tapped, the
// reveal-land's enters-tapped, Mox Diamond's and Heart of Yavimaya's
// graveyard (ADR 0098 Decision 5).
//
// asksItsOwnQuestion answers "may this be fired blind?". This answers
// the other question a window that cannot ask has to answer: which
// branch is the weaker one? For an Optional "may" and a copy selector,
// NOT applying is weaker, so such a window skips them. For these two,
// not applying would be the STRONGER branch — a free untapped
// shockland, a free Mox — so the window runs Replace instead.
//
// Riot (ADR 0109 §10 decision 3) is the third: its Replace is the
// counter, and an entry that cannot ask takes the counter, the default
// entry_controller takes for the same situation. Riot is mandatory —
// one of its two outcomes always happens — so skipping it would be a
// third outcome the card never prints.
func declineIsReplace(e ReplacementEffect) bool {
	return e.EntryLifeCost > 0 || e.EntryCardChoice != nil || e.entryKeyword == KeywordRiot
}

// skipOwnQuestionLocked settles an effect that asks its own question
// in a window that cannot ask it (mustSettleNow, an ordering whose
// affected player has left): it is marked applied, and its Replace
// runs only when that is its weaker branch (declineIsReplace).
//
// Caller must hold g.mu, and must have allocated the once-per-event
// map entry for ev.ID.
func (g *Game) skipOwnQuestionLocked(ev *ReplacementEvent, a activeReplacement) {
	g.replacementsAppliedThisEvent[ev.ID][a.id] = true
	if !declineIsReplace(a.effect) || a.effect.Replace == nil || ev.Canceled {
		return
	}
	if err := a.effect.Replace(ev, g, a.source); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
	}
}

// offerOwnQuestionLocked puts an effect's own question to its player,
// or settles its un-asked branch inline. `handled` is false for an
// effect that asks nothing; `pending` is true when a prompt is now
// queued and the caller must bail.
//
// The ONE dispatcher for the four question shapes, shared by the
// apply-loop's single-applicable arm and ResolveReplacementOrder's
// chosen-order loop. They used to keep two lists, and the second had
// lost the reveal-land branch (ADR 0098 gap 4) the way it had lost the
// "may" before #847.
//
// Caller must hold g.mu.
func (g *Game) offerOwnQuestionLocked(ev *ReplacementEvent, chosen activeReplacement) (pending, handled bool) {
	switch {
	case chosen.effect.EntryController != nil:
		// "Enters under the control of an opponent of your choice" —
		// entry_controller.go (ADR 0102). From ResolveReplacementOrder
		// only when two CR 616.1b effects were ordered against each
		// other; the apply-loop asks it before its cannot-pause skip.
		return g.offerEntryControllerLocked(ev, chosen), true
	case chosen.effect.CopySelector != nil:
		// "You may have this enter as a copy of ..." — copy_choice.go.
		return g.offerCopyChoiceLocked(ev, chosen), true
	case chosen.effect.EntryLifeCost > 0:
		// "As this enters, you may pay N life." — entry_choice.go.
		return g.offerEntryLifePaymentLocked(ev, chosen), true
	case chosen.effect.EntryCardChoice != nil:
		// "You may reveal / discard …", "sacrifice … instead" —
		// entry_card_choice.go.
		return g.offerEntryCardChoiceLocked(ev, chosen), true
	case chosen.effect.entryKeyword == KeywordRiot:
		// Riot's counter or haste (CR 702.136a) — riot.go.
		return g.offerEntryRiotLocked(ev, chosen), true
	case chosen.effect.Optional:
		// A "may" (#847) — the owner decides each time. The two cases
		// that decline it inline — a chooser who has left, an event
		// with nothing to resume it (#359) — live in the helper.
		return g.offerOptionalReplacementLocked(ev, chosen), true
	}
	return false, false
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
	if chosen.effect.EntryController != nil {
		// ADR 0102: mandatory, and nobody can be asked on this path, so
		// the default opponent — never a skip.
		g.settleEntryControllerByDefaultLocked(ev, chosen)
		return
	}
	g.runReplaceLocked(ev, chosen)
}

// skipQuestionsLocked drops the gathered replacements that would ask
// their controller a question — a "may", a shockland's
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
			// ADR 0098 Decision 5: marked applied, and its decline run
			// when the decline IS Replace (a shockland, a reveal-land,
			// a Mox Diamond), so the un-asked branch is the weaker one.
			g.skipOwnQuestionLocked(ev, a)
			continue
		}
		out = append(out, a)
	}
	return out
}

// optionalReplacementResumableLocked reports whether pausing on ev
// for a "may" yes/no prompt has something that can finish the
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
		if eff.commanderZone {
			// #1397: the owner has already answered CR 903.9 for this
			// move — before a cost that cannot pause was paid. A "no"
			// is not a question to ask again; a "yes" is no longer a
			// question at all, so it applies like any mandatory
			// replacement, in the gathered order a settle-now event
			// uses. `eff` is this gather's own copy.
			switch ev.commanderAnswer {
			case commanderZoneDecline:
				continue
			case commanderZoneAccept:
				eff.Optional = false
			}
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
			key := catalogAbilityKeyOf(card)
			reps := CatalogReplacements(key)
			if len(reps) == 0 {
				continue
			}
			for repIdx := range reps {
				if !reps[repIdx].ActiveWhen.Active(*card) {
					continue
				}
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
				if eff.FromGraveyard {
					// Applies from the drawing player's graveyard
					// (dredge), gathered below, never from play.
					continue
				}
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
					identity: catalogReplacementIdentity(eff, key, repIdx, card.Controller),
				})
			}
		}
	}

	// Graveyard replacements: dredge (draw_instead.go).
	out = g.gatherGraveyardReplacementsLocked(ev, applied, out)

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
				if !reps[repIdx].ActiveWhen.Active(entering) {
					continue
				}
				if repIdx >= MaxCatalogReplacementSlots {
					// The same budget the packed catalog IDs are
					// bounded by, for the same reason: slot
					// MaxCatalogReplacementSlots would be the first
					// scoped replacement's ID.
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
			// ADR 0102 decision 6, CR 614.12: once an entering permanent
			// has been given a copy to enter as (CR 616.1c, a Clone's
			// answer), which effects apply is judged "as it would exist
			// on the battlefield" — as the copy. So the COPIED card's
			// own entry replacements are gathered too: a Clone copying
			// Abby, Merciless Soldier enters under the control of an
			// opponent. Their IDs sit one stride above the card's own,
			// so a copied slot 0 is never mistaken for the Clone's
			// already-applied selector. A copied copy selector is not
			// gathered: the copy has been chosen, and asking again would
			// be a second copy of one entry.
			if ev.EntersAsCopyOf != nil {
				copied := entering
				copied.applyCopy(*ev.EntersAsCopyOf, Card{})
				ckey := CatalogKey(copied)
				if ckey != key {
					creps := CatalogReplacements(ckey)
					for repIdx := range creps {
						if repIdx >= MaxCatalogReplacementSlots {
							break
						}
						id := selfReplacementIDBase + MaxCatalogReplacementSlots + ReplacementEffectID(repIdx)
						if applied[id] {
							continue
						}
						eff := creps[repIdx]
						if eff.CopySelector != nil || !eventKindMatches(eff.Watches, ev.Kind) {
							continue
						}
						src := copied
						if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, &src) {
							continue
						}
						out = append(out, activeReplacement{
							effect:   eff,
							source:   &src,
							id:       id,
							identity: replacementIdentity{card: ckey, slot: repIdx, controller: src.Controller},
						})
					}
				}
			}
		}
	}

	// Keyword entry replacements — riot and unleash (ADR 0109 §10),
	// one per instance the CR 614.12 look-ahead finds on the entering
	// permanent as it would exist on the battlefield. Outside the block
	// above because a keyword needs no catalog entry: the deck importer
	// stamps a printed riot on the card itself.
	out = g.gatherEntryKeywordReplacementsLocked(ev, applied, out)

	// Scoped replacements — the replacement mods of ScopedEffect
	// records (ADR 0041 P8, tier 3b): Fog, a prevention shield, the
	// Whip's redirect, Cosmic Intervention. IDs are minted from each
	// record's Seq, never its position, because the registry can
	// shrink under an open CR 616 prompt. See scoped_replacements.go.
	out = g.gatherScopedReplacementsLocked(ev, applied, out)

	// Test replacements. IDs live in a dedicated range above the
	// catalog + scoped spaces and below the built-in base.
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
	// Scoped replacements (Fog, a prevention shield, …), by Seq.
	if id >= scopedReplacementIDBase {
		return g.scopedReplacementLabelLocked(id), uuid.UUID{}
	}
	// Graveyard replacements (dredge).
	if label, card, ok := g.graveyardReplacementMetaLocked(id); ok {
		return label, card
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
		// and so has no slot to look up here — but a keyword entry
		// replacement is named by its keyword (riot.go).
		if kw := EntryKeywordOfReplacement(id); kw != "" {
			return entryKeywordLabel(kw), uuid.UUID{}
		}
		return "", uuid.UUID{}
	}
	if cardIdx >= len(g.Battlefield.Cards) {
		return "", uuid.UUID{}
	}
	card := &g.Battlefield.Cards[cardIdx]
	reps := CatalogReplacements(catalogAbilityKeyOf(card))
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
		// CR 701.9a. A discard is a move out of the hand, but what a
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
		// CR 701.17a. EventMill is the post-event twin and it fires per
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
