package game

import (
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// MinPlayers and MaxPlayers bound a legal game. Commander allows up to
// 10 players per the MTG comprehensive rules, but cmd_and_ctrl is
// scoped to 4-player games per PLAN.md §1, and that bound gates the
// table layout work in S05 and the Commander damage grid in S10.
const (
	MinPlayers = 2
	MaxPlayers = 4
)

// State is the lifecycle state of a game.
type State string

const (
	// StateLobby is the pre-start state: players are joining, decks
	// are being imported. Only AddPlayer and Start are valid.
	StateLobby State = "lobby"

	// StateActive is an ongoing game. Zone mutations and turn advances
	// are legal; AddPlayer is not.
	StateActive State = "active"

	// StateEnded is a finished game. No mutations are legal; the game
	// is retained for replay or inspection.
	StateEnded State = "ended"
)

// Game is the complete server-side state of one Commander table. A
// Game is safe for concurrent access: every public method takes the
// internal rwmutex. Callers outside this package never touch fields
// directly — the exported methods are the whole surface. Read-only
// consumers (like the protocol view builder) should use ReadSnapshot.
//
// Zero-value Game is not valid; use NewGame.
type Game struct {
	ID        uuid.UUID
	CreatedAt time.Time
	State     State

	// Outcome is the result of an ended game — who won, or a draw, and
	// why (ADR 0057 Decision 5, game_end.go). Nil while the game is
	// active, and nil for a game ended by End() with no result (an
	// abandoned table). Written only by endGameLocked.
	Outcome *GameOutcome

	// ActiveSeatLeftPending is set when the ACTIVE player loses the
	// game by an effect in the middle of a resolution (ADR 0057
	// Decision 3). They leave at once, but the game-over check and the
	// turn rotation wait for the next SBA loss pass, which consumes
	// the flag: beginning the next turn must never run inside a
	// resolving callback (ADR 0059 Decision 6). Plain data, carried by
	// Clone and the snapshot.
	ActiveSeatLeftPending bool

	// Seats is the ordered list of players. Index matches Turn.ActiveSeat.
	Seats []*Player

	// Shared zones. Owner is uuid.Nil.
	//
	// Battlefield holds every permanent the game currently treats as
	// existing. That is NOT quite "every permanent on the
	// battlefield": a phased-out permanent is still on the
	// battlefield by CR 702.26d and is held in PhasedOut below, out of
	// this slice, so that CR 702.26b's "treated as though it does not
	// exist" is true of every walk over it without any walk having to
	// ask. See ADR 0084.
	Battlefield *Zone
	Stack       *Zone
	Exile       *Zone

	// PhasedOut holds the permanents that are phased out (CR 702.26).
	//
	// NOT A ZONE IN THE CR 400 SENSE, despite the type: phasing is not
	// a zone change (CR 702.26d) and these permanents are still on the
	// battlefield as far as the rules are concerned. The Zone type is
	// reused so that cloneZone, snapshotZone / restoreZone and
	// viewOfZone work on it unchanged; findCardZoneLocked deliberately
	// does not look here, and nothing routes a card in or out of it
	// but phaseOutLocked / phaseInLocked in phasing.go.
	//
	// Three readers, and each names the rule it is implementing: the
	// untap step's CR 502.1 turn-based action, the CR 800.4a sweep in
	// leave_game.go, and the wire's `phased_out` zone. Added in S46
	// (#1199, ADR 0084).
	PhasedOut *Zone

	// Turn cursor, meaningful only when State == StateActive.
	Turn Turn

	// MulligansOpen is true between Start and the moment all seated
	// players have called KeepHand. While open, the client is
	// expected to surface a keep / mulligan dialog and gate the
	// "real" game UI behind the player's commitment. Added in S08.
	MulligansOpen bool

	// Monarch is the player ID currently designated as the monarch
	// (CR 724). uuid.Nil means "no monarch currently". Handed out by
	// the set_monarch action or by a card effect; from #375 onward the
	// engine then enforces the designation's two inherent triggered
	// abilities — the monarch's end-step draw and the transfer to
	// whoever deals combat damage to them — in monarch.go. Added in
	// S10.
	Monarch uuid.UUID

	// Initiative is the player ID currently designated as having taken
	// the initiative (Commander Legends: Battle for Baldur's Gate
	// mechanic; player ventures into the Undercity at the start of
	// their upkeep). uuid.Nil means "unassigned". Sandbox-only marker;
	// venturing is resolved manually until rules graft work lands.
	// Added in S10.
	Initiative uuid.UUID

	// Settings is the table's configuration: the undo budget and
	// scope, starting life, the commander damage threshold, bot pace
	// and the spawn switch (ADR 0075 §2.2, settings.go). NewGame sets
	// DefaultTableSettings; UpdateSettings changes it, in the lobby or
	// mid-game. Carried by Clone and the snapshot, and deliberately
	// NOT rolled back by RestoreFrom — an undo must not be able to
	// undo the setting that limits undos. Replaced the S11 UndoLimit
	// field in S35 (#1032).
	Settings TableSettings

	// StartingSeat is the seat index that took the first turn. Set in
	// Start() to the active seat at game start. Used by the StepDraw
	// auto-action to skip the starting player's turn-1 draw per
	// CR 103.8a — in a two-player game only, because CR 103.8c has
	// nobody skip it in any other multiplayer game.
	// Replays predating S13 default to seat 0 on decode,
	// which matches the only seat games started on before this field
	// existed. Added in S13.
	StartingSeat int

	// StackMeta is the per-stack-item metadata map (S13.1). Keyed by
	// the item's ID — for spells this equals the underlying Card's
	// InstanceID (the card itself lives in Game.Stack); for abilities
	// it's a synthetic UUID minted at activation/announce time. The
	// dual storage (Game.Stack for the spell card slice, StackMeta
	// for everything else) keeps existing zone plumbing untouched
	// while giving cast/resolve code a place to attach announce-time
	// choices (targets, modes, X, distribution, hold-priority,
	// split-second).
	StackMeta map[uuid.UUID]*StackItem

	// PendingTriggers is the APNAP queue of triggered abilities
	// waiting to hit the stack (CR 603.3b). Drained on the next
	// priority-grant boundary in active-player-non-active-player
	// order. Per-controller chosen ordering of multiple triggers
	// from the same player is captured at announce time. Added in
	// S13.1.
	PendingTriggers []*StackItem

	// DelayedTriggers is the queue of CR 603.7 delayed triggered
	// abilities — "at the beginning of the next end step, return
	// that card to the battlefield". Each entry sits dormant until
	// the turn cursor ENTERS the step it names, at which point
	// runStepEntryHooksLocked drains it onto PendingTriggers like
	// any other trigger, so it uses the stack and can be responded
	// to. Unlike TriggeredAbility these are not declared on a
	// catalog Spec and need no source permanent in play: the card
	// that created one is usually already in a graveyard. See
	// delayed.go. Added in S22.
	DelayedTriggers []*DelayedTrigger

	// SplitSecondActive mirrors "any item on the stack has
	// SplitSecond set" (CR 702.61). While true, cast_spell and
	// activate_ability return ErrSplitSecondActive. Mana abilities
	// and special actions stay legal. Recomputed every time the
	// stack changes (cast, counter, resolve). Added in S13.1.
	SplitSecondActive bool

	// LoyaltyActivatedThisTurn flags planeswalkers whose loyalty
	// abilities have already been activated this turn (CR 606.3).
	// Keyed by the planeswalker's instance ID. Cleared on
	// Turn.advance to a new turn (i.e. when ActiveSeat changes).
	// Added in S13.1.
	LoyaltyActivatedThisTurn map[uuid.UUID]bool

	// SpellsCastThisTurn tallies, per player, the spells that player
	// has cast this turn (CR 700.7-style "first spell each turn"
	// bookkeeping; also the storm count once storm ships). Bumped
	// in CastSpell before EventCast is emitted, so a "whenever a
	// player casts their first noncreature spell each turn" trigger
	// reads Noncreature == 1 for the spell that just fired the
	// event. Cleared on Turn.advance to a new turn. Keyed by player
	// ID. Added in S19 sub-PR 6.
	SpellsCastThisTurn map[uuid.UUID]CastTally

	// ForetoldThisTurn tallies, per player, how many cards that
	// player has foretold this turn (CR 702.143a's special action,
	// not a cast) — SpellsCastThisTurn's shape, one zone over.
	// Bumped in foretellLocked once the card has actually landed in
	// exile, so a special action refused before payment never counts.
	// Read by SpecialActionManaCostForEffect's cost modifiers —
	// Ranar the Ever-Watchful's "The first card you foretell each
	// turn costs {0} to foretell" (#1319) is FirstForetellEachTurn()
	// asking whether this count is still zero. Cleared on
	// Turn.advance to a new turn, alongside SpellsCastThisTurn.
	ForetoldThisTurn map[uuid.UUID]int

	// LandsPlayedThisTurn counts, per player, the lands that player
	// has played this turn via CastSpell's land branch. Cleared on
	// Turn.advance to a new turn. Keyed by player ID. Added in S31
	// sub-PR 1.
	//
	// #500: this is now the ENFORCED side of CR 305.2, not just
	// bookkeeping for the legal-move enumerator. CastSpell refuses a
	// land play once the count reaches the player's allowance —
	// EffectiveLandDropsLocked, see land_drops.go.
	LandsPlayedThisTurn map[uuid.UUID]int

	// ExtraLandDropsThisTurn holds one-turn grants of additional land
	// plays ("you may play an additional land this turn"), per player.
	// Added to the player's base allowance and to any static grant
	// from a controlled permanent; see land_drops.go. Written by
	// GrantAdditionalLandPlayForEffect and cleared on Turn.advance to
	// a new turn with the rest of the per-turn state. Added in #500.
	ExtraLandDropsThisTurn map[uuid.UUID]int

	// DrawnThisTurn records, per player, the card instance IDs that
	// player has DRAWN this turn, in draw order. Appended by
	// actuallyDrawCardLocked — the one place a card crosses from
	// library to hand as a draw — and cleared on Turn.advance to a
	// new turn, alongside the other per-turn tallies above.
	//
	// It is a list of IDs rather than a count because the cards
	// themselves are the thing that gets read: Sylvan Library's "choose
	// two cards in your hand drawn this turn" is a prompt whose
	// candidate set IS this slice, intersected with the hand. A tally
	// could not name a card.
	//
	// Entries are NOT removed when a card leaves hand. A card drawn and
	// then discarded was still drawn this turn, and every reader
	// intersects with the zone it cares about anyway, so pruning here
	// would cost a hook on every zone move to buy nothing.
	DrawnThisTurn map[uuid.UUID][]uuid.UUID

	// TurnTally counts what has happened this turn — life gained and
	// lost, cards drawn, creatures died, tokens, sacrifices, landfall,
	// attacks, combat damage to players, and per-ability resolution /
	// trigger counts — bumped by turnTallyListener as the events fire
	// and reset on Turn.advance to a new turn. See turn_tally.go
	// (#586).
	TurnTally TurnTally

	// Activations counts what has been ACTIVATED — CR 602.2b's
	// announcement — per (object, printed ability), in two scopes at
	// once: `Ever`, which is never reset and is what exhaust reads,
	// and `Turn`, which is emptied on the turn advance beside
	// TurnTally. Written at the one place a non-mana activation is
	// paid for (ActivateCatalogAbility). See activation_tally.go
	// (#1181).
	Activations ActivationTally

	// LoopNotice is the CR 726 loop breaker's flag: set when the
	// same triggered ability has resolved LoopThreshold times this
	// turn with no player decision in between, nil otherwise. While
	// it is set, AUTOMATIC passing is suspended — the client's
	// autopass toggle and the bot runner both hold — so the humans
	// get priority back with the loop's trigger still on the stack.
	// Cleared by the next cast, activation, answered prompt or
	// combat declaration. See loop_breaker.go and ADR 0055 (#628).
	LoopNotice *LoopNotice

	// LoopThreshold overrides DefaultLoopThreshold for this game.
	// Zero (the production value) means "use the default". It is a
	// field rather than a package global so a test can trip the
	// breaker in three resolutions without every other test in the
	// tree sharing the setting; set it before Start.
	LoopThreshold int

	// DiscardPending is the cleanup-step pause map (S13.4): keys
	// are player IDs that need to discard, values are the count
	// each player must discard. Set at cleanup-step entry by
	// populateDiscardPendingLocked, for the active player only
	// (CR 514.1), when their hand is over their maximum hand size;
	// cleared per-player by the discard_selection action. The
	// cleanup auto-advance is blocked while this map is non-empty so
	// the cursor pauses for player input. Added in S13.4.
	//
	// It IS cleanup-only (#651). Until then DiscardChoiceForEffect
	// (Mind Rot, looting) wrote here too, and the two obligations do
	// not have the same shape: nothing outside cleanup waits on this
	// map, and populateDiscardPendingLocked RESETS it on cleanup
	// entry, which erased any effect discard still owed. An effect's
	// discard is part of the resolving effect (CR 608.2c) and is now
	// a PendingChoice; this map is the CR 514.1 turn-based action and
	// nothing else.
	DiscardPending map[uuid.UUID]int

	// Promises is the directed per-pair "I owe you" promise-token count
	// keyed by `from→to` pairs. Politics scaffold — players use it as
	// a visual reminder of informal deals ("I owe Alice 2 favours").
	// Set semantics via SetPromise; cleared to zero by setting count=0
	// (the entry stays in the map for rendering simplicity, server
	// trims trailing zeros at view-time). Added in S10.
	Promises map[PromiseKey]int

	// Vote is the currently open council's-dilemma / vote-on-an-issue,
	// or nil when no vote is active. The wire view exposes it as a
	// small object so every player sees the same prompt and tally.
	// Added in S10.
	Vote *Vote

	// PendingChoices is the S14 generic "someone needs to pick"
	// queue. Populated by effect cards that defer a decision
	// (Thoughtseize → caster picks from target's revealed hand).
	// Distinct from DiscardPending (S13.4 cleanup-only). Drained
	// by resolve_choice actions. See pending_choice.go.
	PendingChoices []*PendingChoice

	// Events is the append-only per-game event log. Every rules-
	// visible mutation calls EmitEvent, producing one or more
	// entries here in-order. Consumers are S14+ card-effect
	// primitives (sub-PR 3 onwards) and S19's triggered-ability
	// harvester. Clone deep-copies this slice; RestoreFrom replaces
	// it. See events.go for the tagged-struct shape. Added in S14
	// sub-PR 1.
	Events []Event

	// eventSeq is the monotonic counter stamped into Event.Seq on
	// each EmitEvent. Starts at 0; first emitted event has Seq == 1.
	// Survives Clone / RestoreFrom so reconstructed games continue
	// the sequence rather than restarting at 1.
	eventSeq uint64

	// eventLogGen names the history Events holds (#1401): bumped by
	// RestoreFrom, which replaces the log with a shorter one that the
	// next emit regrows under the same Seq values. Never copied — it
	// belongs to this *Game, not to the history — so it only moves
	// forward. See projection_cache.go.
	eventLogGen uint64

	// logProjection is the public log's fold state between views
	// (#1401). Opaque to this package; see projection_cache.go. Not
	// cloned, not restored, not snapshotted: a fresh *Game starts with
	// an empty slot and the first view refolds from event 0.
	logProjection ProjectionCache

	// eventBatch is the monotonic counter stamped into Event.Batch on
	// each EmitEvent: the identity of the run of events the engine is
	// emitting as ONE occurrence (CR 603.2c). It advances at exactly
	// one boundary — beginEventBatchLocked, called when a stack item
	// begins to resolve and when the turn cursor enters a new step.
	// See event_batch.go (#829).
	//
	// oncePerBatchFired records, per TallyKey(source, ability key),
	// the batch a OncePerBatch ability last fired for. That is the
	// whole of the "whenever one or more …" guard: an ability whose
	// key is already recorded against the live batch declines, and
	// anything else fires.
	//
	// The ability key carries a SECOND dimension when the printed
	// clause quantifies over something (#784, CR 603.2c) — "deal
	// combat damage to A PLAYER" is once per player, not once per
	// step — appended by TriggeredAbility.BatchKey. No extra state:
	// the same map, a longer key. See event_batch.go.
	//
	// Both survive Clone / RestoreFrom together, for the reason
	// TurnTally does: an undo that rewound the counter but kept the
	// marks (or the reverse) would either double-fire a trigger or
	// swallow one.
	eventBatch        uint64
	oncePerBatchFired map[string]uint64

	// announcedBlocks and blockedAttackers are what this combat's
	// block declaration has produced (#830, #715). announcedBlocks
	// maps blocker -> the attacker its EventBlock named;
	// blockedAttackers is the set of attackers that are BLOCKED.
	//
	// blockedAttackers is the CR 509.1h state: "a creature remains
	// blocked even if all the creatures blocking it are removed from
	// combat". It is written in exactly one place —
	// commitBlockDeclarationLocked, the block declaration's lock-in —
	// and read by both combat damage steps, which never ask the live
	// battlefield whether an attacker is blocked (#715: they used to,
	// so an attacker whose chump blocker died hit the player). The
	// same set is the CR 506.4 announcement guard, because an
	// attacker becomes blocked exactly once: an EventBecomesBlocked
	// is emitted when, and only when, an attacker is added here.
	//
	// The maps exist because the declaration is announced at LOCK-IN
	// and the sandbox lets the defender keep clicking afterwards: the
	// commit emits events only for what has changed since, so a
	// second blocker added to an already-blocked attacker announces
	// its own block and no second "becomes blocked", and a blocker
	// re-pointed after the lock-in does not re-announce the attacker
	// it left. Both are cleared by clearCombatLocked, which is also
	// what clears BlockingTarget — they are one combat's bookkeeping.
	// removeFromCombatLocked drops one permanent's rows when an
	// effect takes it out of combat (CR 506.4).
	//
	// Carried by Clone / RestoreFrom together for the reason
	// eventBatch and oncePerBatchFired are: an undo that rewound the
	// declaration but kept the announcements would swallow the
	// re-done trigger, and the reverse would double-fire it — and an
	// undo that dropped the blocked state would hand a blocked
	// attacker's damage to the defending player.
	announcedBlocks  map[uuid.UUID]uuid.UUID
	blockedAttackers map[uuid.UUID]bool

	// announcedAttacks is the same bookkeeping for the ATTACK
	// declaration (#859, attackers.go): the creatures that have had
	// their one EventAttack this combat. Presence, not a pairing,
	// because CR 508.1 declares a creature as an attacker once — a
	// re-point is a correction to a declaration that is already
	// announced, never a second attack. It doubles as the marker for
	// a permanent PUT onto the battlefield attacking (CR 506.3c),
	// which is never announced at all.
	//
	// Cleared by clearCombatLocked, and carried by Clone /
	// RestoreFrom and the persisted snapshot, for the reasons the two
	// block maps above are.
	announcedAttacks map[uuid.UUID]bool

	// attackDefenders is each attacker's DEFENDING PLAYER as it was
	// when its attack was last pointed (#1364, ADR 0045 Decision 35):
	// written by setAttackTargetLocked at the declaration, at an entry
	// "attacking" (CR 506.3c) and at a reselect (CR 508.7), and never
	// rewritten when the planeswalker or battle it attacks leaves.
	//
	// It exists for CR 506.4c's "it may be blocked": once the attacked
	// permanent is gone the live resolution (defendingPlayerForAttack-
	// Locked) has nobody to name, and CR 802.2a says the defending
	// player is still the one it was attacking before the permanent
	// was removed from combat. Read ONLY by the block path through
	// defendingPlayerForAttackerLocked — damage still goes nowhere
	// (CR 510.1b).
	//
	// Cleared with the rest of combat, dropped per object with
	// announcedAttacks, and carried by Clone / RestoreFrom and the
	// persisted snapshot for the same reasons.
	attackDefenders map[uuid.UUID]uuid.UUID

	// blocksDeclared is the set of DEFENDING PLAYERS whose CR 509.1
	// block declaration is complete this combat (#1279, ADR 0045
	// Decision 38, block_completion.go). Written only by
	// completeBlockDeclarationLocked — at the step's entry for a
	// defender with no legal block, when the defender passes priority
	// or sends finish_blocks, and for everyone still pending as the
	// cursor leaves the step — and it is what tells "this defender has
	// not declared yet" from "this defender declared no blocks", which
	// an empty blockedAttackers cannot.
	//
	// Cleared with the rest of combat, and carried by Clone /
	// RestoreFrom and the persisted snapshot for the reason
	// announcedBlocks is: an undo across the declaration that kept the
	// bit would make an attacker read unblocked to ninjutsu before the
	// defender had chosen, and one that dropped it would reopen a
	// declaration whose triggers have already fired.
	blocksDeclared map[uuid.UUID]bool

	// firstStrikeStepParticipants is THIS combat's CR 510.4 / 702.7c
	// participation record: the attacking and blocking creatures that
	// had first strike or double strike as the FIRST combat damage
	// step began (#716). It is recorded once, at that step's entry,
	// and both damage steps read it instead of re-reading keywords:
	//
	//   - the first-strike step deals damage for exactly this set;
	//   - the regular step deals damage for every combatant NOT in
	//     it, plus the ones that have double strike right now.
	//
	// Re-reading the keywords in the second step is the bug #716
	// reports. Between the two steps there is a real priority window
	// (#717), so a lord granting first strike can die to first-strike
	// damage, or be Murdered in the window: the creature it pumped
	// has already dealt its damage and must not deal it again, and
	// one that GAINS first strike in the window is owed its ordinary
	// damage in the second step rather than a first-strike hit it
	// missed.
	//
	// Empty means the first-strike step did not happen, so every
	// combatant deals damage in the single combat damage step — which
	// is also what the zero value gives a combat that never had one.
	// Cleared by clearCombatLocked alongside the declarations it
	// describes, and carried by Clone / RestoreFrom and the persisted
	// snapshot for the reason announcedBlocks is: the window between
	// the steps is a priority window, so an undo or a restore can land
	// inside it, and a record that was dropped there would let a
	// first-striker hit twice.
	firstStrikeStepParticipants map[uuid.UUID]bool

	// Listeners is the per-game event subscriber list. Populated by
	// RegisterListener; walked by notifyListenersLocked under the
	// write lock. S14 ships the registry infrastructure with zero
	// production listeners; S16 added layerVersionBump; S19 sub-PR 1
	// adds triggerHarvester (the auto-fire dispatcher). Shallow-
	// copied on Clone / RestoreFrom (listeners are process-lifetime
	// singletons). Added in S14 sub-PR 1.
	Listeners []Listener

	// lastKnownBattlefield holds CR 603.10 last-known-information
	// snapshots for cards leaving the battlefield. Populated by
	// snapshotLKILocked just before each battlefield-exit MoveCard;
	// read + cleared by the S19 trigger harvester when EventLTB
	// fires. Empty in steady state — entries live for the duration
	// of one mutation that emits LTB. See triggers.go. Added in S19
	// sub-PR 1.
	lastKnownBattlefield map[uuid.UUID]Characteristic

	// lastKnownTriggerIdentity retains the small part of a leaving card that
	// CR 400.7 cleanup removes before the move-form event is harvested. It is
	// paired with lastKnownBattlefield and cleared at the same boundary.
	lastKnownTriggerIdentity map[uuid.UUID]triggerIdentityLKI

	// lastKnownCounters is lastKnownBattlefield's sibling for a card's
	// COUNTERS (#1218): Characteristic deliberately excludes Counters
	// ("belongs to other engine subsystems"; see characteristic.go),
	// and MoveCard's battlefield-exit cleanup zeroes Card.Counters
	// (zone.go) before an LTB trigger — or a bystander watching the
	// event, The Ozolith's "if it had counters on it" — ever sees it.
	// Populated by snapshotLKILocked from the live counters just
	// before that zeroing; read by LastKnownCountersForEffect; kept
	// on the GAME rather than the card for the same reason the other
	// two are — the departing Card value is about to be overwritten
	// out from under whatever holds a copy of it. Cleared at the same
	// two boundaries lastKnownBattlefield is.
	lastKnownCounters map[uuid.UUID]map[string]int

	// lastKnownStack is CR 608.2h last-known information for spells
	// that left the stack WITHOUT resolving this turn — countered,
	// returned, moved by hand — keyed by the spell's instance ID and
	// holding the card and its stack item as they last stood. Read
	// only by CopyLastKnownSpellForEffect, the copy entry point for
	// effects that name a spell without targeting it (storm). Cleared
	// at the turn boundary. See stack_lki.go (#1255).
	lastKnownStack map[uuid.UUID]lastKnownSpell

	// lastKnownPermanents is CR 608.2h last-known information for
	// permanents that left the battlefield this turn, one entry per
	// departed OBJECT (instance ID + ObjectEpoch). Unlike
	// lastKnownBattlefield it outlives the exit's trigger harvest, so
	// an ability that resolves later can read the power a creature had
	// as it left. Written by battlefieldExitLocked, read by
	// PermanentForEffect, cleared at the turn boundary. See
	// permanent_lki.go (#1379).
	lastKnownPermanents map[uuid.UUID][]PermanentInfo

	// simultaneousExit holds copies of the permanents currently
	// leaving the battlefield as ONE event — a board wipe, or one
	// state-based-action sweep. Non-empty only for the duration of
	// that sweep; the trigger harvester reads it so a watcher that
	// died earlier in the same wipe still sees its neighbours die
	// (CR 700.4 / 603.10). See simultaneous.go. Transient by
	// construction, so Clone / RestoreFrom do not carry it: no
	// snapshot is ever taken mid-sweep. Added in S23.
	simultaneousExit []Card

	// enteringTokens holds the tokens whose CR 614 battlefield-entry
	// window is open and which are therefore in NO zone yet: minted,
	// not pushed. A card entering the battlefield sits in the zone it
	// is leaving while its entry replacements are consulted; a token
	// has no such zone (CR 111.1 — it is created on the battlefield),
	// and an entry replacement still has to be able to read it.
	// Urabrask the Hidden asks whether the entering permanent is a
	// creature an opponent controls, and it asks through
	// LookupCardForEffect, which is why that function looks here when
	// no zone holds the card.
	//
	// Non-empty only for the duration of one token's entry — normally
	// a few statements, and across a prompt when that entry pauses on
	// a CR 616 ordering question. Clone copies it so an undo across
	// such a prompt still has the token; the persisted snapshot does
	// not carry it, exactly as it does not carry the once-per-event
	// map the same paused window is holding. See token_create.go.
	enteringTokens []Card

	// resolving is the stack item whose resolution is in flight, plus
	// the card it was (CR 608.2n last-known information), held for
	// exactly one event batch — see resolving_item.go (#920).
	//
	// resolveTopOfStackLocked removes an item from StackMeta BEFORE it
	// runs, which is right: the item is no longer on the stack and
	// nothing may target it. But CR 707.10 lets a resolving spell
	// create a copy of ITSELF (Chain of Vapor, Chain of Smog), and
	// CopySpellForEffect finds its source through StackMeta, so with
	// the entry already gone the copy leg found nothing. This slot is
	// where it looks instead.
	//
	// Its lifetime is the OCCURRENCE, not the function call: the copy
	// decision is a prompt, so it is answered after the resolution
	// has returned and the spell has already reached the graveyard.
	// beginEventBatchLocked clears it, which makes the terminal
	// boundary exactly the next resolution or the next step — the
	// same boundary #829 already draws, and one no open prompt can be
	// crossed by, because a resolution-time prompt blocks the table.
	//
	// Clone carries it so an undo across that prompt still has the
	// source; the persisted snapshot drops it, for the reason
	// enteringTokens is dropped — it is non-empty between actions only
	// for a resolution paused on a prompt, and that prompt's resume
	// frame is already counted in ContinuationCensus.ChoiceResumeFrames.
	resolving *resolvingItem

	// resolutionOpen says a stack item has begun to resolve and the
	// CR 704.3 boundary after it has not run yet (#1289). Set with the
	// resolving slot, cleared by runStateChecksLocked once no prompt the
	// resolution queued is still open. While it is set, a prompt queued is stamped
	// PendingChoice.midResolution, and an open stamped prompt holds the
	// state-based actions and the trigger drain. See
	// resolution_pause.go.
	//
	// resolutionDepth counts the resolution functions currently on the
	// Go stack. Inside one, the item has not finished resolving
	// whatever it has queued, so the boundary is held there too.
	//
	// Clone and the persisted snapshot carry resolutionOpen; the depth
	// is zero between actions by construction.
	resolutionOpen  bool
	resolutionDepth int

	// The game's randomness: a secret key plus per-stream draw
	// counters for the current turn (ADR 0054 Decision 2). Every
	// random draw goes through randForLocked in rng.go, which derives
	// a fresh generator from these three; nothing holds a *rand.Rand.
	//
	// rngKey is minted at Start (from crypto/rand, or read from the
	// caller's seeded source) and never changes. The zero key means
	// "not minted yet"; the first draw mints one. It is PERSISTED in
	// the engine snapshot and must never reach a GameView: anyone
	// holding it could predict every future draw.
	//
	// rngCounters counts the draws each stream has taken this turn,
	// and rngTurn is the turn index they belong to
	// (rngTurnIndexLocked). Clone and
	// RestoreFrom carry all three, so undo REWINDS randomness: the
	// same action after an undo gets the same draw (ADR 0054
	// Decision 4).
	//
	// Not concurrency-safe on their own; every read and write goes
	// through a path already holding g.mu in write mode.
	rngKey      [32]byte
	rngCounters map[string]uint64
	rngTurn     int

	// sourceOrdinals makes an object's random-effect stream stable across
	// independently seeded runs. Deck cards use class 0 (seat/index); objects
	// created while playing use class 1 and the monotonically increasing
	// sourceOrdinalNext. See ADR 0054 addendum.
	sourceOrdinals    map[uuid.UUID]uint64
	sourceOrdinalNext uint64

	// layerVersion is the S16 continuous-effect-engine invalidation
	// counter. Bumped by listeners on every event that could change
	// which static abilities are active or what they apply to —
	// battlefield zone changes, control changes, counter changes,
	// step advance. Sub-PR 1 ships the field; sub-PR 2 wires the
	// listener bumps; sub-PR 3 wires actual recompute work.
	// lastResolvedVersion mirrors the most recent version the
	// recompute pass has caught up to; the snapshot path no-ops when
	// they're equal. Both atomic so the staleness CHECK can run
	// under the read lock without racing a listener bump — the
	// recompute itself takes the write lock (see ReadSnapshot).
	// Read via .Load(), bumped via .Add(1) / .Store(). Added in
	// S16 sub-PR 1.
	layerVersion        atomic.Uint64
	lastResolvedVersion atomic.Uint64

	// recomputeCount instruments the layer engine for the fast-path
	// regression test (exit criterion #10): two consecutive snapshot
	// reads with no events between invoke RecomputeLayersLocked
	// exactly once. Test-only; production callers ignore. Atomic
	// for the same reason as the version counters.
	recomputeCount atomic.Uint64

	// BuiltinReplacements is the S17 replacement-effect registry
	// populated at NewGame. Today: just the commander-zone
	// replacement (refactored from S13.1's inline
	// applyCommanderZoneReplacementLocked). Future built-ins
	// register by appending here before Start. Catalog replacements
	// flow through CatalogReplacements (see effect_hooks.go) and
	// are NOT listed here. See replacements.go +
	// builtin_replacements.go. Added in S17 sub-PR 2.
	BuiltinReplacements []ReplacementEffect

	// TurnScopedReplacements is the per-turn replacement slot —
	// effects registered here live until the current turn's
	// StepCleanup, then are cleared. Used by spells that create
	// transient replacement effects (Fog's "prevent all combat
	// damage this turn", future "until end of turn" damage
	// prevention cards). Distinct from BuiltinReplacements (game-
	// lifetime) and catalog replacements (battlefield-presence-
	// gated via source card's AppliesTo). Added in S17 sub-PR 5.
	TurnScopedReplacements []ReplacementEffect

	// TurnScopedBlockRules is the UNTIL-END-OF-TURN slot for the
	// CR 509.1b block rules of #750 — Gingerbrute's "target creature
	// can't block this turn"-shaped effects, and any other block
	// restriction a spell or ability creates for the turn rather than
	// a permanent printing. Read by forEachBlockRuleLocked alongside
	// the battlefield walk, so a turn-scoped rule is a pair check and
	// a count bound exactly as a catalog one is.
	//
	// Emptied wholesale by ClearTurnScopedBlockRulesLocked in the
	// cleanup sweep, next to ClearTurnScopedReplacementsLocked. The
	// registry carries no turn stamp because it holds nothing that
	// outlasts a turn: no card found while drafting ADR 0045's
	// addendum prints a block rule with a longer duration, and one
	// that arrives would use #755's duration model rather than a
	// second one here.
	//
	// Closures, so the snapshot cannot carry it: counted in
	// ContinuationCensus.TurnScopedBlockRules and marked `dropped` in
	// the drift test, exactly like TurnScopedReplacements.
	// See ADR 0045 addendum Decision 11 and block_rules.go.
	TurnScopedBlockRules []BlockRule

	// ScopedStatics is the CONTINUOUS-EFFECT slot for effects whose
	// lifetime is a duration rather than a battlefield source: Giant
	// Growth's +3/+3, Overrun's mass pump and trample grant, Act of
	// Treason's theft, Agent of Treachery's. Consulted by
	// activeStaticAbilitiesLocked alongside the battlefield walk and
	// swept by the one duration sweep (CR 611.2). See
	// scoped_statics.go and duration.go. Added in S32 as
	// TurnScopedStatics; renamed in S38 when it stopped being
	// turn-scoped (ADR 0063).
	ScopedStatics []ScopedStatic

	// ScopedEffects is the same slot's DATA twin (ADR 0041 phase 3,
	// #1497): a continuous effect from a resolution recorded as an
	// affected set, a list of operations from a closed vocabulary and
	// a duration, so the snapshot carries it and a game holding one is
	// still a restore point. Adapted into the layer pass beside
	// ScopedStatics and swept by the same duration sweep. See
	// scoped_effects.go.
	ScopedEffects []ScopedEffect

	// testReplacements is the test-only replacement injection slot
	// populated by RegisterReplacementForTest. Unexported so
	// production code has no path to it. Walked after built-ins +
	// catalog in gatherActiveReplacementsLocked. Added in S17
	// sub-PR 2.
	testReplacements []ReplacementEffect

	// replacementsAppliedThisEvent maps ReplacementEvent.ID → set
	// of ReplacementEffectIDs that have already fired for that
	// event (CR 614.5 / 616.1 once-per-event tracking). Scope is
	// per-pipeline-call: pipeline functions defer-clear the entry
	// for ev.ID at their outermost frame so CR 616 prompt pauses
	// don't lose the entry mid-event. See replacements.go.
	// Added in S17 sub-PR 2.
	//
	// So between actions it holds an entry only for an event paused
	// on a prompt, and that entry is part of the prompt's state:
	// Clone deep-copies it and RestoreFrom puts it back, or an undo
	// into the open prompt would replay the answer without the marks
	// and fire an already-applied effect twice (#808).
	replacementsAppliedThisEvent map[ReplacementEventID]map[ReplacementEffectID]bool

	// nextReplacementEventID mints the per-event keys stored in
	// replacementsAppliedThisEvent. Atomic so pipeline functions
	// can mint without promoting the lock (in practice they
	// already hold g.mu, but atomic is defensive). Added in S17
	// sub-PR 2.
	nextReplacementEventID atomic.Uint64

	// promptRuns holds the prompted runs in flight — sacrifices
	// (#1019) and discards (#1027) alike — keyed by the id each of a
	// run's prompts carries on PendingChoice.promptRun
	// (prompt_run.go). One run is one printed instruction; the entry
	// lives from the moment its prompts are queued until the last of
	// them has settled and its continuation has run.
	//
	// ONE registry for both verbs, because the keys are freshly
	// minted uuids and "which prompt is a leg of which run" is one
	// question: a second map would be a second place for a new verb's
	// clone, restore and census wiring to be forgotten.
	//
	// So between actions it holds an entry only for an instruction
	// paused on a prompt, and that entry is part of the prompt's
	// state: Clone deep-copies it and RestoreFrom puts it back, for
	// exactly the reason replacementsAppliedThisEvent above does — an
	// undo into the open prompt would otherwise replay the answer
	// against a counter that had already been decremented and pay out
	// a seat early.
	promptRuns map[uuid.UUID]*promptRun

	mu sync.RWMutex
}

// NewGame constructs a game in the lobby state with a fresh ID and
// empty shared zones.
func NewGame() *Game {
	g := &Game{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		State:       StateLobby,
		Battlefield: newZone(ZoneBattlefield, uuid.Nil),
		Stack:       newZone(ZoneStack, uuid.Nil),
		Exile:       newZone(ZoneExile, uuid.Nil),
		PhasedOut:   newZone(ZonePhasedOut, uuid.Nil),
		Settings:    DefaultTableSettings(),
	}
	// S16 sub-PR 2: install the layer-engine invalidation listener.
	// Bumps g.layerVersion on the events that change which static
	// abilities are active (battlefield zone moves) or what they
	// apply to (counter changes). See layer_listener.go.
	// #586: the per-turn tally. Registered FIRST so it reads a dying
	// creature's last-known characteristics before the harvester
	// drops them, and so a trigger's AppliesTo asking "already this
	// turn?" during the same event sees the count as it stood before
	// that event. See turn_tally.go.
	g.Listeners = append(g.Listeners, turnTallyListener{})
	g.Listeners = append(g.Listeners, layerVersionBump{})
	// S19 sub-PR 1: install the auto-fire trigger dispatcher. Walks
	// the battlefield (and the LKI map for LTB events) on every
	// emit, queues matching catalog-declared TriggeredAbility
	// entries onto PendingTriggers. See triggers.go.
	g.Listeners = append(g.Listeners, triggerHarvester{})
	// #375: the two inherent triggered abilities of the monarch
	// (CR 724.2) have no source card for the harvester above to find
	// them on, so they ride the listener registry instead. Registered
	// AFTER the harvester so that when a card trigger and a monarch
	// trigger watch the same event, the card's lands on
	// PendingTriggers first — CR 603.3b reorders anything that
	// actually matters. See monarch.go.
	g.Listeners = append(g.Listeners, monarchTriggers{})
	// S17 sub-PR 2: install the CR 903.9 commander-zone built-in
	// replacement. Refactored from S13.1's inline
	// applyCommanderZoneReplacementLocked. See
	// builtin_replacements.go.
	g.BuiltinReplacements = append(g.BuiltinReplacements,
		commanderZoneReplacement,
		regenerationShieldReplacement,
		// #662, CR 702.16e. Order in this slice is not precedence —
		// the protection built-in declares Preemptive, which is what
		// puts it ahead of every other applicable replacement and
		// keeps a charged prevention shield unspent (ADR 0072 §4).
		protectionPreventsDamageReplacement,
		// #1200, CR 119.7 / CR 119.8 (ADR 0085, life_lock.go). Also
		// Preemptive, and for the same reasons one event kind over:
		// there is only one answer once an amount replacement has
		// been applied, and a life PAYMENT cannot pause to be asked.
		// It watches the LIFE event; the damage half of the rule is
		// in applyResolvedDamageToPlayerLocked, because damage does
		// not fire this window (CR 120.3).
		lifeTotalCantChangeReplacement,
	)
	return g
}

// AddPlayer seats a new player at the table with the given name and
// starting decklist. The deck is placed into the player's library in
// the order supplied; commanders (Card.IsCommander == true) are routed
// to the command zone instead.
//
// Returns the newly-seated player. AddPlayer is only valid in the
// lobby state; calling it after Start returns ErrGameNotInLobby.
func (g *Game) AddPlayer(name string, deck []Card) (*Player, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if len(deck) == 0 {
		return nil, ErrEmptyDeck
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return nil, ErrGameNotInLobby
	}
	if len(g.Seats) >= MaxPlayers {
		return nil, ErrGameFull
	}

	seat := len(g.Seats)
	p := newPlayer(name, seat)
	p.Life = g.Settings.StartingLife

	// Stamp every card with its owner (overwriting whatever the caller
	// set) and route commanders vs. library cards.
	for i, c := range deck {
		c.Owner = p.ID
		c.Controller = p.ID
		g.setDeckSourceOrdinalLocked(c.InstanceID, seat, i)
		if c.IsCommander {
			p.Command.PushTop(c)
		} else {
			p.Library.PushTop(c)
		}
	}

	g.Seats = append(g.Seats, p)
	return p, nil
}

// ReplaceDeck swaps a seated player's library + command zone with
// the supplied deck. Only valid while the game is still in the lobby
// state — once Start runs, deck mutations would let a player top-
// deck arbitrary cards mid-game.
//
// Every supplied card is re-stamped with the owner's ID; callers
// shouldn't pre-populate Owner / Controller. Commanders (IsCommander)
// route to the command zone, everything else to the library in
// supplied order. Shuffle happens on Start.
//
// Returns ErrPlayerNotFound if playerID isn't seated and
// ErrGameNotInLobby if the game has already started.
func (g *Game) ReplaceDeck(playerID uuid.UUID, deck []Card) error {
	if len(deck) == 0 {
		return ErrEmptyDeck
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return ErrGameNotInLobby
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}

	// Reset the zones in-place. Reusing the existing *Zone pointers
	// keeps any external references (logs, snapshots mid-serialise)
	// from dangling.
	p.Library.Cards = p.Library.Cards[:0]
	p.Command.Cards = p.Command.Cards[:0]

	for i, c := range deck {
		c.Owner = p.ID
		c.Controller = p.ID
		g.setDeckSourceOrdinalLocked(c.InstanceID, p.Seat, i)
		if c.IsCommander {
			p.Command.PushTop(c)
		} else {
			p.Library.PushTop(c)
		}
	}
	p.DeckImported = true
	return nil
}

// Start transitions the game from lobby to active with seat 0 as the starting
// player and shuffles each player's library. It is the stable fixture/replay
// entry point retained for callers that already chose a starting seat.
// Production game creation uses StartWithFirstPlayerRoll instead.
//
// The RNG argument no longer IS the game's source; it seeds the
// game's secret key (ADR 0054 Decision 2). A caller-supplied source
// gives a key read from it, so a test that starts from a seeded
// *rand.Rand draws the same shuffles on every run — and, unlike
// before, that game is persistable and rewindable like any other.
//
// Pass nil to have the engine mint its key from crypto/rand. That is
// what production does. Start(nil) keeps a key that is already set
// (SetRNGKeyForTest), so a test can fix the key before starting.
//
// Returns ErrNotEnoughPlayers if fewer than MinPlayers are seated,
// and ErrGameAlreadyStarted if the game is not in lobby.
func (g *Game) Start(r *rand.Rand) error {
	if r == nil {
		return g.start(nil, false)
	}
	key := rngKeyFrom(r)
	return g.start(&key, false)
}

// StartWithFirstPlayerRoll is Start plus the real-table pregame procedure:
// every seat rolls a d20, tied leaders reroll, and the winner takes turn 1.
// The rolls use the same seeded, persisted RNG as card effects and are emitted
// to the public log. Production callers should use this entry point.
func (g *Game) StartWithFirstPlayerRoll(r *rand.Rand) error {
	if r == nil {
		return g.start(nil, true)
	}
	key := rngKeyFrom(r)
	return g.start(&key, true)
}

// StartWithSource is Start on a caller-supplied PCG source. It used
// to exist because only a PCG's position could be persisted; with a
// keyed RNG every source is equivalent, and it is kept so callers do
// not change. The key is read from the source exactly as Start reads
// it from a *rand.Rand.
//
// Pass nil to get the same crypto-minted key Start(nil) mints.
func (g *Game) StartWithSource(src *rand.PCG) error {
	if src == nil {
		return g.start(nil, false)
	}
	key := rngKeyFrom(src)
	return g.start(&key, false)
}

// start is the shared body. key is the game's RNG key, or nil to keep a key
// already set and otherwise mint one. rollForFirst selects whether this caller
// already chose seat 0 or wants the table's public opening roll.
func (g *Game) start(key *[32]byte, rollForFirst bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return ErrGameAlreadyStarted
	}
	if len(g.Seats) < MinPlayers {
		return ErrNotEnoughPlayers
	}

	switch {
	case key != nil:
		g.rngKey = *key
		g.rngCounters = nil
	case g.rngKey == ([32]byte{}):
		g.rngKey = mintRNGKey()
		g.rngCounters = nil
	}
	// No "<= 0 means default" rewrite here any more (ADR 0075 §2.2):
	// NewGame sets the defaults, and a limit of 0 set in the lobby is
	// a table that wants no undos.
	// S13.5: collect every seated player's ID so command-zone
	// initialisation can mark all knowers in one pass.
	allSeatedIDs := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		allSeatedIDs = append(allSeatedIDs, p.ID)
	}
	for _, p := range g.Seats {
		p.UndosRemaining = g.Settings.UndoLimit
		p.Life = g.Settings.StartingLife
		// The opening shuffle is the pre-game turn index's first draw
		// on this seat's shuffle stream (the cursor is set to turn 1
		// below).
		p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
		// Deal an opening hand of 7. If the library is too short to
		// satisfy 7 (a malformed deck), stop early — the partial hand
		// is still valid and tests can use small decks.
		for i := 0; i < OpeningHandSize; i++ {
			c, err := p.Library.PopTop()
			if err != nil {
				break
			}
			p.Hand.PushTop(c)
		}
		// S13.5: opening-hand cards are known to their owner only.
		// Library cards have no knowers (post-shuffle order is
		// unknown to everyone). Command-zone cards are public — every
		// seated player sees them per CR 400.2.
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(p.ID)
		}
		for i := range p.Command.Cards {
			p.Command.Cards[i].AddKnowersAll(allSeatedIDs)
		}
	}
	g.StartingSeat = 0
	if rollForFirst {
		g.StartingSeat = g.rollStartingSeatLocked()
	}
	g.Turn = newStartingTurn(g.StartingSeat)
	// The one turn that does not begin through the rotation seam
	// still counts as a turn begun (ADR 0063 Decision 3).
	g.noteTurnBegunLocked(g.StartingSeat)
	g.State = StateActive
	g.MulligansOpen = true
	// Step entry hooks (auto-untap, auto-draw, etc.) intentionally do
	// NOT fire at Start — the cursor sits idle on Untap until the
	// mulligan window closes (KeepHand triggers them when MulligansOpen
	// flips false). This keeps the lobby/keep-hand UX from auto-drawing
	// before players have committed to their opening hand.
	return nil
}

// OpeningHandSize is the number of cards each player draws when the
// game starts. The mulligan flow can take a player back to the same
// count for redraws (simplified — no London bottom-N penalty yet).
const OpeningHandSize = 7

// DefaultUndoLimit is the default per-player undo budget refreshed
// each turn (TableSettings.UndoLimit). One is intentionally tight —
// undo is for "I clicked the wrong card", not for re-litigating
// turns. The table can change it through UpdateSettings. Added in S11.
const DefaultUndoLimit = 1

// End transitions the game to the ended state. Idempotent: calling
// End on an already-ended game is a no-op.
//
// An admin closing a table, not a result: Outcome stays nil, and the
// wire has no winner for it (ADR 0057 Decision 5). Every rules-driven
// end goes through endGameLocked instead.
func (g *Game) End() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateEnded {
		// CR 702.143f, #658. The same sweep endGameLocked
		// runs, on the other door out of an active game — an admin
		// ending the table, and every caller that ends it outright.
		g.revealForetoldAtGameEndLocked()
	}
	g.State = StateEnded
}

// AdvanceStep is the sandbox's skip-ahead: "pass priority until this
// step ends." With an empty stack that is one step of the cursor, the
// way it always was. With something ON the stack it is the passes the
// rules require first — CR 117.4, a step or phase ends only once every
// player has passed in succession with an empty stack — so whatever
// the step owed resolves INSIDE the step instead of after the next
// one's turn-based actions (#914). After the cleanup step the cursor
// wraps to the next seat's untap step, the turn sequence increments,
// and the round increments on a rotation.
// Returns the Turn the call ends on.
//
// Side effects on step transitions:
//   - Entering first_strike_damage: the first-strike damage pass
//     runs (CR 510.4). The step exists only when a combatant has
//     first or double strike as it would begin; otherwise the cursor
//     walks through it without entering it.
//   - Entering combat_damage: ResolveCombatDamage runs (S08 —
//     auto-applies unblocked attacker damage to defending players'
//     life totals). Combat declarations (AttackingTarget /
//     BlockingTarget) are intentionally left in place.
//   - LEAVING end_combat: clearCombatLocked wipes the combat
//     declarations (CR 511.3 — creatures are removed from combat as
//     that step ENDS, #785). Everything that happens in the end of
//     combat step, "at end of combat" triggers included, still sees
//     the attackers and blockers; the client's combat arrows come
//     down when the cursor reaches postcombat_main.
//
// Returns ErrGameNotActive if the game is not in the active state,
// and a *ChoicePendingError while a blocking prompt is open (#730).
// A prompt raised by a resolution the drive itself caused is NOT an
// error: the drive stops there with the cursor where it is, so the
// table sees the prompt against the board that raised it.
func (g *Game) AdvanceStep() (Turn, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return Turn{}, ErrGameNotActive
	}
	// #830 / CR 509.2a, and #859 / CR 508.2: leaving a step completes
	// whatever turn-based action was staged in it. A combat
	// declaration still staged here — attackers or blockers — is
	// locked in BEFORE the cursor moves, so its triggers are
	// harvested inside the declaring step and off the final
	// assignment. Placed above the prompt gate on purpose: an
	// optional declaration trigger (Grazilaxx's "you may return it",
	// Legion Loyalty's myriad) queues its yes/no here, and the gate
	// then holds the cursor until it is answered rather than walking
	// the table past it. No-op whenever nothing is staged.
	//
	// #1279: and every defender still declaring blockers completes
	// here, as whatever they have staged — AdvanceStep is the cursor
	// leaving the step (completion point 4, block_completion.go).
	blocksCompleted := g.completeAllBlockDeclarationsLocked()
	if blocksCompleted || g.blockDeclarationPendingLocked() || g.attackDeclarationPendingLocked() {
		g.runStateChecksLocked()
	}
	// #730: an unanswered prompt gates the table. Checked before the
	// cursor moves, so a choice queued by THIS advance's step-entry
	// hooks (a CR 616 ordering pause, say) is not mistaken for one
	// the table walked past. See choice_gate.go.
	if c := g.blockingChoiceLocked(); c != nil {
		return g.Turn, choicePendingErrorLocked(c)
	}
	// #914 / CR 117.4: a step does not end while the stack has
	// anything on it. Drive the table's passes until it is empty —
	// each resolution followed by SBAs, the trigger drain and another
	// priority round, exactly as ordinary play does it — and only
	// then move the cursor.
	switch g.driveStepToEndLocked() {
	case driveStepEnded:
		// The last pass wrapped with an empty stack and ended the
		// step itself. That path runs the same entry hooks this one
		// does; finish with the state checks below so AdvanceStep's
		// postcondition is the same however the step ended.
		g.runStateChecksLocked()
		return g.Turn, nil
	case driveHalted:
		// A prompt, a loop notice or the game ending stopped the
		// drive. The cursor stays in the step that still owes
		// something; no error, because the resolutions that got this
		// far are real and the caller must see them.
		return g.Turn, nil
	}
	g.advanceCursorLocked()
	g.runStepEntryHooksLocked()
	// CR 117.5 / 704.3: SBAs fire whenever a player would get
	// priority. AdvanceStep lands on a priority-granting step (Untap
	// and Cleanup auto-advance through their hooks), so this is such
	// a boundary — run the SBA + APNAP-trigger-drain loop to catch
	// 0-life losses, lethal damage, 0-loyalty planeswalkers, etc.
	// that accumulated during the prior step without a priority pass.
	g.runStateChecksLocked()
	return g.Turn, nil
}

// driveResult says how the CR 117.4 drive in AdvanceStep ended.
type driveResult int

const (
	// driveCursorMayMove: nothing is on the stack in the step the
	// call started in, so the cursor moves the way it always has.
	// Also the answer when the step grants no priority at all
	// (untap, cleanup) and there is therefore nothing to drive —
	// the cursor still moves, so a stray stack item cannot wedge a
	// step that no pass_priority could unstick either.
	driveCursorMayMove driveResult = iota
	// driveStepEnded: the passes emptied the stack and the wrap ended
	// the step on its own. The cursor has already moved.
	driveStepEnded
	// driveHalted: a blocking prompt, a CR 726 loop notice, or the
	// game ending stopped the drive with the step still owing
	// something. The cursor has not moved.
	driveHalted
)

// maxAdvanceStepPasses bounds the CR 117.4 drive. One resolution
// costs up to one pass per seat, so this is ~256 resolutions at a
// four-player table — far past anything a step legitimately owes, and
// past the CR 726 loop breaker's own threshold, which stops the drive
// long before this does. Hitting it halts the drive with the cursor
// where it stands; the caller clicks again.
const maxAdvanceStepPasses = 1024

// driveStepToEndLocked passes priority around the table for the
// caller until the step's stack is empty — CR 117.4's "all players
// pass in succession with the stack empty", which is what the sandbox
// skip-ahead has to mean if it is not to walk past what the step owes
// (#914).
//
// It reuses PassPriority's body rather than resolving anything
// itself: one priority engine, so the drive resolves, runs SBAs,
// drains triggers and hands priority back in exactly the order a
// table of humans clicking "next" would.
//
// It stops on the three things that stop automatic passing anywhere
// else in the engine: a prompt addressed to somebody (#730 /
// ADR 0018 §6), a CR 726 loop notice (ADR 0055 — checked AFTER a
// pass, so a standing notice still lets one manual nudge through the
// way the client's "next" button does), and the game ending under a
// resolution.
//
// Caller must hold g.mu.
func (g *Game) driveStepToEndLocked() driveResult {
	start := g.Turn
	for i := 0; i < maxAdvanceStepPasses; i++ {
		// A resolution that queued a prompt stops the drive here
		// rather than inside passPriorityLocked, so the halt is a
		// state the caller can broadcast rather than an error over a
		// board that has already changed. Ahead of the stack check
		// because the LAST resolution of the drive is the one most
		// likely to ask a question: an empty stack with a prompt
		// standing on it is still a cursor #730 must not move.
		if g.blockingChoiceLocked() != nil {
			return driveHalted
		}
		if !g.stackHasItemsLocked() {
			return driveCursorMayMove
		}
		if err := g.passPriorityLocked(); err != nil {
			if errors.Is(err, ErrNoPriority) {
				return driveCursorMayMove
			}
			return driveHalted
		}
		if g.Turn.Step != start.Step ||
			g.Turn.Seq != start.Seq ||
			g.Turn.ActiveSeat != start.ActiveSeat {
			return driveStepEnded
		}
		if g.LoopNotice != nil {
			return driveHalted
		}
	}
	return driveHalted
}

// advanceCursorLocked moves the step cursor forward by one — the
// single seam every step transition goes through. Past cleanup it
// hands over to beginNextTurnLocked (rotation.go), which skips seats
// that have left the game (CR 800.4a / 800.4k) and runs the
// turn-began hook so the per-turn caches clear.
//
// Both halves fix bugs the S31 bot fuzzer found on its first run:
// Turn.advance rotated into eliminated seats, handing priority to a
// player who could not act (humans had been escaping with
// advance_step), and the cleanup hook's wrap never reset the per-turn
// caches, so SpellsCastThisTurn / LoyaltyActivatedThisTurn survived
// every ordinary turn change. Caller must hold g.mu.
func (g *Game) advanceCursorLocked() {
	// CR 511.3: "as soon as the end of combat step ends, all
	// creatures, battles and planeswalkers are removed from combat."
	// This is where that step ends — the one seam every step
	// transition goes through — so this is where combat is cleared
	// (#785). It used to happen on ENTRY to end_combat, which took
	// every attacker and blocker out of combat for the whole of the
	// step the rules keep them in: Aetherize, Settle the Wreckage and
	// Aetherspouts cast there found nothing, Desert had no legal
	// target in the only step it can be activated, and "activate only
	// if you control an attacking creature" was false there. The "at
	// end of combat" triggers are unaffected: CR 511.2 fires them as
	// the step BEGINS, from the step-entry hook's announcement, with
	// the creatures still in combat.
	//
	// The manual ClearCombat verb, PassTurn and the eliminated-seat
	// rotation keep their own calls — a turn that ends early never
	// reaches this seam.
	if g.Turn.Step == StepEndCombat {
		g.clearCombatLocked()
	}
	// #829: entering a step is one of the two points where play moves
	// on, so the events this step emits are a new occurrence — first
	// strike damage and regular damage are two batches, as in paper.
	// See event_batch.go.
	g.beginEventBatchLocked()
	if g.Turn.Step == StepCleanup {
		g.beginNextTurnLocked()
		return
	}
	g.Turn = g.Turn.advance(len(g.Seats), g.StartingSeat)
}

// CardsDrawnThisTurnFor returns the instance IDs playerID has drawn
// this turn, in draw order, filtered to the cards still in that
// player's hand.
//
// The filter is the part that matters: "cards in your hand drawn this
// turn" (Sylvan Library) is the printed wording, and a card drawn and
// then discarded is no longer eligible. Returns a fresh slice, so a
// caller can hold it across a prompt without aliasing engine state.
//
// Caller must hold g.mu.
func (g *Game) CardsDrawnThisTurnFor(playerID uuid.UUID) []uuid.UUID {
	if len(g.DrawnThisTurn) == 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, id := range g.DrawnThisTurn[playerID] {
		if p.Hand.Contains(id) {
			out = append(out, id)
		}
	}
	return out
}

// LandsPlayedThisTurnFor returns how many lands p has played this
// turn (zero when none). Read-only helper for the legal-move
// enumerator. Caller must hold g.mu.
func (g *Game) LandsPlayedThisTurnFor(playerID uuid.UUID) int {
	if g.LandsPlayedThisTurn == nil {
		return 0
	}
	return g.LandsPlayedThisTurn[playerID]
}

// CastTally is one player's per-turn spell count. Total counts every
// spell; Noncreature counts those without the creature type (per the
// card's printed type line at cast time). Added in S19 sub-PR 6.
type CastTally struct {
	Total       int
	Noncreature int
}

// CastTallyFor returns p's tally for the current turn (zero value
// when they haven't cast anything). Read-only helper for triggered
// abilities' AppliesTo predicates. Caller must hold g.mu.
func (g *Game) CastTallyFor(playerID uuid.UUID) CastTally {
	if g.SpellsCastThisTurn == nil {
		return CastTally{}
	}
	return g.SpellsCastThisTurn[playerID]
}

// ForetoldCountThisTurn returns how many cards p has foretold this
// turn (zero when they haven't foretold anything) — CastTallyFor's
// shape, one zone over. Read by SpecialAction cost modifiers deciding
// whether "the first card you foretell each turn" still applies.
// Caller must hold g.mu.
func (g *Game) ForetoldCountThisTurn(playerID uuid.UUID) int {
	if g.ForetoldThisTurn == nil {
		return 0
	}
	return g.ForetoldThisTurn[playerID]
}

// stepExistsLocked reports whether the step the cursor has landed on
// is one THIS turn actually has. A step that does not exist is walked
// through by runStepEntryHooksLocked without being entered: nothing
// announces, no turn-based action runs, and no player receives
// priority in it.
//
// One step answers false today. CR 506.1 / 510.4: a combat phase has
// TWO combat damage steps, the first of them for first strike, only
// "if at least one attacking or blocking creature has first strike or
// double strike as the combat damage step begins". The cursor is at
// that moment right now, so the check is the live board — which is
// also why it is a predicate here rather than a flag set when blockers
// were declared: a creature can gain or lose first strike during the
// declare-blockers step's own priority window.
//
// The regular combat damage step always exists (every combat has one),
// and so does every other step in turnSequence.
//
// Caller must hold g.mu.
func (g *Game) stepExistsLocked(s Step) bool {
	if s != StepFirstStrikeDamage {
		return true
	}
	return len(g.firstStrikeStepParticipantSetLocked()) > 0
}

// runStepEntryHooksLocked dispatches the per-step side effects that
// fire on entering certain steps. Called from every code path that
// changes Turn.Step (AdvanceStep, PassPriority's wrap-and-advance
// branch, KeepHand when the mulligan window closes) so the auto-
// resolve / auto-clear hooks fire regardless of which mutation
// produced the step transition.
//
// S13 introduced auto turn-based actions on StepUntap, StepDraw, and
// StepCleanup. Untap and Cleanup do not grant priority, so after
// firing their action the hook recurses through g.Turn.advance to
// drive the cursor on to the next priority-granting step. A single
// client-visible step advance from StepEnd therefore walks End →
// Cleanup → next seat's Untap → next seat's Upkeep inside one write
// lock.
//
// Side effects by step:
//   - StepUntap (S13): the CR 502 turn-based action — untap the
//     permanents the active seat controls (plus any an
//     UntapStepPermission adds) and clear their summoning sickness;
//     refresh their per-turn undo budget; auto-advance because Untap
//     grants no priority.
//   - StepDraw (S13): draw 1 for the active seat, except when the
//     starting player would draw on turn 1 of a TWO-player game
//     (CR 103.8a skip; CR 103.8c has no one skip at a larger table).
//   - StepCombatDamage: auto-resolve unblocked attacker damage.
//   - StepEnd (S22): emit EventBeginEndStep so "at the beginning of
//     your end step" triggers fire. The delayed-trigger drain that
//     runs just before the switch covers "at the beginning of the
//     next end step" for the same boundary.
//   - StepEndCombat: nothing. The step has no turn-based action
//     (CR 511.1), and the removal from combat happens as it ENDS
//     (CR 511.3), from advanceCursorLocked — not here (#785).
//   - StepCleanup (S13): the CR 514.1 hand-size discard pauses the
//     cursor here (S13.4); the CR 514.2 sweep runs; then cleanup.go's
//     one exit either ends the turn (CR 514.3) or gives the active
//     player priority in this step and comes back for a second
//     cleanup step (CR 514.3a, #661).
//
// Caller must hold g.mu.
func (g *Game) runStepEntryHooksLocked() {
	// The step entry reads the layered board before anything else:
	// the skip-step replacement window below, the untap step's set
	// (CatalogAbilityKey on an UntapStepPermission source, the
	// layer-2 Controller) and every trigger the step announcement
	// harvests. The previous step can leave the cache stale with no
	// priority boundary in between — cleanup's "until end of turn"
	// sweep and the turn wrap both bump the layer version and then
	// recurse straight into the next seat's untap and upkeep, and the
	// untap step's own untaps bump it again on the way into upkeep.
	// Fast-path no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	// #717 / CR 506.1: not every turn has every step. A step this
	// turn does not have never begins — no announcement, no turn-based
	// action, no priority — so the cursor walks straight through it,
	// which is the same move a step CANCELLED by a replacement makes
	// below (CR 500.11). Checked before the replacement window because
	// a step that does not exist is not a step anything can replace.
	if !g.stepExistsLocked(g.Turn.Step) {
		g.advanceCursorLocked()
		g.runStepEntryHooksLocked()
		return
	}
	// S17 sub-PR 2: step-transition replacement hook. Stasis
	// cancels StepUntap; Necropotence's "skip your draw step"
	// cancels StepDraw. A cancelled step is a SKIPPED step
	// (CR 500.11): advance past it and recurse through this hook so
	// the cursor lands on the next step's entry.
	//
	// #710: the window can PAUSE. Two skip-step effects under one
	// controller (Necropotence + Yawgmoth's Bargain; Stasis plus any
	// second untap skip) are two applicable replacements on one
	// event, and CR 616 asks the affected player to order them. This
	// used to fall straight through to the step body, so the prompt
	// was queued and the draw happened anyway — the opposite of what
	// both cards say. The rest of the step entry now lives in
	// finishStepEntryLocked, which the resume in
	// applyResolvedReplacementEventLocked calls with the settled
	// answer. (Two pure cancels no longer prompt at all — see
	// ReplacementEffect.PureCancel — but the pause has to be correct
	// for the mixed case regardless.)
	stepEv := &ReplacementEvent{
		Kind:               RepEventStepTransition,
		StepTransitionStep: g.Turn.Step,
		StepTransitionSeat: g.Turn.ActiveSeat,
	}
	out, err := g.applyReplacementsLocked(stepEv)
	if errors.Is(err, errReplacementPending) {
		// A CR 616 ordering (or "may") prompt is queued.
		// The step does NOT begin: nothing below runs, nothing
		// announces, and the tracking-map entry stays alive for the
		// resume, which clears it.
		return
	}
	defer g.clearReplacementEventLocked(stepEv.ID)
	// Canceled events come back as (nil, nil) from
	// applyReplacementsLocked — check err==nil + out==nil as
	// the cancel signal, plus the belt-and-braces out.Canceled
	// for any intermediate path that returns the event.
	g.finishStepEntryLocked(err == nil && (out == nil || out.Canceled))
}

// finishStepEntryLocked is the second half of a step entry: what
// happens once the CR 614 replacement window over the transition has
// settled. `canceled` is the window's verdict — true means the step
// is skipped (CR 500.11), so the cursor advances past it and the next
// step's entry hook runs instead of this step's turn-based action.
//
// Split out of runStepEntryHooksLocked in #710 for the same reason
// applyResolvedDamageLocked was split out of the damage entry points
// in #694: the CR 616 resume has to finish the transition with
// exactly the code the unpaused path runs, and the only way to
// guarantee that is for there to be one copy of it. Everything below
// the cancel branch is verbatim what ran inline before.
//
// Caller must hold g.mu.
func (g *Game) finishStepEntryLocked(canceled bool) {
	// Also reached from a CR 616 prompt's resume, which is not a path
	// through runStepEntryHooksLocked's recompute.
	g.RecomputeLayersIfStaleLocked()
	if canceled {
		// Step canceled — advance past and recurse so the
		// cursor hits the next step's entry hook.
		g.advanceCursorLocked()
		g.runStepEntryHooksLocked()
		return
	}
	// CR 106.4: every player's mana pool empties at the end of each
	// step / phase. We model this by clearing pools at the START of
	// the next step's entry — equivalent net effect, and centralised
	// here so every step transition (including the auto-advance
	// recursion through Untap → Upkeep and Cleanup → next-Untap)
	// triggers the clear. Cheap to call on already-empty pools.
	g.emptyAllManaPoolsLocked()
	// The mulligan window holds the cursor at Untap and KeepHand
	// re-runs this hook once per keep, so until the window closes the
	// step has not really begun: nothing below announces, fires or
	// untaps. (The per-step cases used to carry this guard each; #588
	// hoisted it so EventStepBegan is emitted exactly once per step
	// that begins and the harvester can trust it.)
	if g.MulligansOpen && (g.Turn.Step == StepUntap || g.Turn.Step == StepUpkeep) {
		return
	}
	// S31 sub-PR 0: announce the step for the public game log, and
	// since #588 for the trigger harvester too. Placed after the
	// skip-step replacement window so a cancelled step never
	// announces, and before the per-step turn-based actions so the
	// entries they produce read as happening INSIDE this step.
	if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) && g.Seats[g.Turn.ActiveSeat] != nil {
		g.EmitEvent(Event{
			Kind:   EventStepBegan,
			Actor:  g.Seats[g.Turn.ActiveSeat].ID,
			Amount: g.Turn.Seq,
			Round:  g.Turn.Round,
			Label:  string(g.Turn.Step),
			Step:   g.Turn.Step,
		})
	}
	// S22: CR 603.7 delayed triggered abilities fire on ENTRY to the
	// step they name. Draining here — before the per-step turn-based
	// actions below — is what makes "the NEXT end step" work without
	// any created-this-step bookkeeping: an ability scheduled during
	// an end step is queued after this hook has already run for that
	// step, so it waits for the following one. See delayed.go.
	g.fireDelayedTriggersLocked(g.Turn.Step)
	switch g.Turn.Step {
	case StepDeclareBlockers:
		// #1279 / CR 509.1: the declaration is taken as the step
		// begins. A defending player with no legal block has nothing
		// to decide, so their declaration — none — is complete now,
		// and nothing that waits on it (ninjutsu, "attacks and isn't
		// blocked", the bot's block grace) waits for a pass that means
		// nothing. A defender with a block to make stays pending; see
		// block_completion.go for how they finish.
		g.autoCompleteBlockDeclarationsLocked()
	case StepPrecombatMain:
		// S27 / CR 714.3: "after your draw step, put a lore counter
		// on each Saga you control" is a turn-based action performed
		// as the precombat main phase begins. Precombat main grants
		// priority, so there is no auto-advance — the chapter
		// triggers this queues are drained onto the stack by the
		// caller's drainPendingTriggersAPNAPLocked / runStateChecks,
		// which is the same boundary every other turn-based action
		// uses.
		g.advanceSagasForActiveSeatLocked()
		// S30: announce the precombat main phase so "at the beginning
		// of your precombat main phase" triggers auto-fire through
		// the harvester. Same shape as the upkeep and end-step
		// announcements, and it follows the Saga advance for the same
		// reason CR 714.3 puts that first: the turn-based action
		// happens as the phase begins, and the triggers that watch
		// the phase go on the stack above whatever it queued.
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			g.EmitEvent(Event{
				Kind:  EventBeginPrecombatMain,
				Actor: g.Seats[g.Turn.ActiveSeat].ID,
			})
		}
	case StepEnd:
		// S22: announce the end step so "at the beginning of your
		// end step" triggers auto-fire through the harvester. The
		// end step grants priority, so no auto-advance.
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			g.EmitEvent(Event{
				Kind:  EventBeginEndStep,
				Actor: g.Seats[g.Turn.ActiveSeat].ID,
			})
		}
	case StepUpkeep:
		// S19 sub-PR 5: announce the upkeep so "at the beginning of
		// your upkeep" triggers auto-fire through the harvester.
		// Upkeep grants priority, so no auto-advance — just emit and
		// fall through.
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			g.EmitEvent(Event{
				Kind:  EventBeginUpkeep,
				Actor: g.Seats[g.Turn.ActiveSeat].ID,
			})
		}
	case StepUntap:
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			// CR 502.1-502.3, in untap.go: the active seat's
			// permanents untap, plus
			// whatever an UntapStepPermission (Seedborn Muse)
			// adds. Each untap emits EventUntapCard, so the
			// harvester sees "whenever a permanent becomes
			// untapped" here — the triggers it queues wait on
			// PendingTriggers and go on the stack at the next
			// priority boundary, which is the upkeep, exactly as
			// CR 502.4 requires of a step that grants none.
			//
			// #826: the step can PAUSE. CR 502.3's "the active
			// player determines which permanents they control will
			// untap" is a real decision under a Winter Orb cap or
			// a "you may choose not to untap" clause, and the
			// prompt stops the step halfway with nothing untapped
			// and the cursor unmoved. Its continuation calls
			// exitUntapStepLocked, which is what this case would
			// have called. See untap_choice.go (ADR 0070).
			if g.performUntapStepLocked(g.Turn.ActiveSeat) {
				return
			}
		}
		// Untap grants no priority; recurse into the next step.
		g.exitUntapStepLocked()
	case StepDraw:
		if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
			return
		}
		// CR 103.8a: in a two-player game the player who takes the
		// first turn skips their draw step on turn 1. CR 103.8c: in
		// every other multiplayer game nobody skips it, so a 3- or
		// 4-player table's starting seat draws like everyone else.
		// (CR 103.8b's Two-Headed Giant case does not apply — the
		// engine has no team format.) Subsequent turns are normal.
		if g.Turn.Seq == 1 && g.Turn.ActiveSeat == g.StartingSeat &&
			g.startingPlayerCountLocked() == 2 {
			return
		}
		// Best-effort: an empty library on auto-draw is not a hard
		// error here (losing from drawing from an empty library is a
		// SBA that lands in S13.1). The drawCardLocked helper surfaces
		// ErrZoneEmpty, which we swallow so the cursor keeps moving —
		// the losing player will be caught by the SBA once it exists.
		activeSeatID := g.Seats[g.Turn.ActiveSeat].ID
		_ = g.drawCardLocked(activeSeatID)
		// #1315 / CR 504.1 widened: "you draw a card during each
		// opponent's draw step" (Teferi, Who Slows the Sunset's
		// emblem) is part of the SAME turn-based action as the active
		// player's own draw — no stack, no priority window in
		// between — so it runs here, before the trigger-eligible
		// announcement below. This is untap.go's Seedborn Muse
		// widening one turn-based action over; see draw_step.go.
		for _, extra := range g.activeDrawStepPermissionsLocked(activeSeatID) {
			drawer := extra.permission.Drawer(g, extra.source)
			n := extra.permission.N
			if n <= 0 {
				n = 1
			}
			for i := 0; i < n; i++ {
				_ = g.drawCardLocked(drawer)
			}
		}
		// S22: announce the draw step so "at the beginning of each
		// player's draw step" triggers auto-fire (Howling Mine).
		// AFTER the draw, per CR 504.1/504.2 — the turn-based draw
		// does not use the stack and goes first; the trigger goes on
		// the stack when priority arrives, which is next.
		g.EmitEvent(Event{
			Kind:  EventBeginDrawStep,
			Actor: g.Seats[g.Turn.ActiveSeat].ID,
		})
	case StepFirstStrikeDamage:
		// CR 510.4: the first combat damage step's turn-based action.
		// It grants priority like any other step, so there is no
		// auto-advance — the triggers this damage causes go on the
		// stack at the boundary resolveFirstStrikeCombatDamageLocked
		// runs, and the active player gets to respond before the
		// second step's damage is dealt (CR 510.3).
		g.resolveFirstStrikeCombatDamageLocked()
	case StepCombatDamage:
		g.resolveCombatDamageLocked()
	case StepCleanup:
		// CR 402.2: build the discard-pending map for any player
		// over their per-player MaxHandSize. The cursor pauses at
		// cleanup until every entry is drained via discard_selection
		// (S13.4 — the interactive discard pathway).
		g.populateDiscardPendingLocked()
		// CR 514.2: marked damage is removed and "until end of turn"
		// and "this turn" effects end. The sweep is its own function
		// (rotation.go) because a turn that ends early — its active
		// player left the game, or the sandbox pass_turn verb — ends
		// through the same code (ADR 0059 Decision 6, #766).
		g.sweepTurnEndLocked()
		// CR 514.3 / 514.3a: the turn ends here — unless a state-based
		// action fired or a trigger is waiting, in which case the
		// active player gets priority in this step and another cleanup
		// step follows. The cursor also sits here, with
		// PriorityHolder=NoPriority, while any player owes a discard;
		// DiscardSelection calls the same exit once the pending map
		// drains. One exit, one decision — see cleanup.go (#661).
		g.exitCleanupStepLocked()
	}
}

// startingPlayerCountLocked reports how many players the game began
// with — the number CR 103.8 keys the turn-1 skip-draw rule off.
//
// Eliminated players deliberately still count. The rule is about the
// game's player count at the start ("a two-player game", CR 103.8a vs
// "all other multiplayer games", CR 103.8c), not about who is still
// alive when the draw step arrives; a concession on turn 1 does not
// retroactively turn a three-player game into a two-player one. Seats
// are only ever added or removed in StateLobby (RemovePlayer refuses
// once the game is active, and a player leaves an active game by
// conceding), so the live seat count still is the count at Start.
//
// Caller must hold g.mu.
func (g *Game) startingPlayerCountLocked() int {
	n := 0
	for _, p := range g.Seats {
		if p != nil {
			n++
		}
	}
	return n
}

// populateDiscardPendingLocked records an over-max discard count
// for the ACTIVE player only when their hand exceeds their
// MaxHandSize. Per CR 514.1 — cleanup-step discard is a turn-based
// action performed only by the active player, not the whole table.
// Eliminated player or NoMaxHandSize (-1) skips the check.
// Caller must hold g.mu.
func (g *Game) populateDiscardPendingLocked() {
	g.DiscardPending = nil
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return
	}
	p := g.Seats[g.Turn.ActiveSeat]
	if p == nil || p.Eliminated {
		return
	}
	// #338: the cap is DERIVED, not read straight off the player.
	// A Thought Vessel on the battlefield lifts it without ever
	// writing to Player.MaxHandSize, so nothing has to be restored
	// when the permanent leaves.
	max := g.EffectiveMaxHandSizeLocked(p)
	if max == NoMaxHandSize {
		return
	}
	over := p.Hand.Size() - max
	if over <= 0 {
		return
	}
	g.DiscardPending = map[uuid.UUID]int{p.ID: over}
}

// CurrentState returns the game's lifecycle state under the read
// lock. Out-of-package pre-checks (lobby join / start / replay
// gating) must use this instead of reading State directly — WS
// actions mutate State under the write lock (e.g. a concede flips it
// to StateEnded), so an unlocked read races.
func (g *Game) CurrentState() State {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.State
}

// WinnerSeat returns the seat of the winner of an ended game, read
// from Game.Outcome (ADR 0057 Decision 5) — which is what makes an
// effect win, where several seats are still standing, name the right
// player. ok is false while the game is not over, for a draw, and for
// an ended game with no single survivor. Read by the lobby to fill
// games.winner_seat (ADR 0051 decision 4).
//
// A game ended with no Outcome — End(), or a snapshot written before
// the engine recorded one — falls back to the one seat left standing,
// which is how the winner was derived before ADR 0057.
func (g *Game) WinnerSeat() (seat int, ok bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.State != StateEnded {
		return 0, false
	}
	if g.Outcome != nil {
		if g.Outcome.Kind != OutcomeWin {
			return 0, false
		}
		if p := g.playerByIDLocked(g.Outcome.Winner); p != nil {
			return p.Seat, true
		}
		return 0, false
	}
	found := -1
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if found >= 0 {
			return 0, false
		}
		found = p.Seat
	}
	if found < 0 {
		return 0, false
	}
	return found, true
}

// ActivePlayer returns the player whose turn it currently is, or nil
// if the game has not yet started.
func (g *Game) ActivePlayer() *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.State != StateActive || len(g.Seats) == 0 {
		return nil
	}
	return g.Seats[g.Turn.ActiveSeat]
}

// PlayerByID looks up a seated player by ID. Returns nil if no player
// with that ID is seated.
func (g *Game) PlayerByID(id uuid.UUID) *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.playerByIDLocked(id)
}

// ControllerOfCard returns the current controller of the card with the
// given instance ID, scanning every zone (battlefield, stack, exile,
// and each player's library / hand / graveyard / command). The second
// return is false when no zone holds the card.
//
// This exists so the actions package can authorize card-instance-
// scoped actions (tap, move_card, add_counter, etc.) by caller without
// reaching into Game internals or holding a write lock — it takes its
// own read lock and the result is a value copy.
func (g *Game) ControllerOfCard(instanceID uuid.UUID) (uuid.UUID, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	z := g.findCardZoneLocked(instanceID)
	if z == nil {
		return uuid.Nil, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == instanceID {
			return c.Controller, true
		}
	}
	return uuid.Nil, false
}

// playerByIDLocked is the unlocked variant of PlayerByID. The caller
// must already hold g.mu (read or write).
func (g *Game) playerByIDLocked(id uuid.UUID) *Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// WithWriteLock runs fn while holding the game's write lock. Used
// by callers (e.g. the room's Undo path) that need to perform a
// multi-field state replacement atomically without exposing every
// individual mutator. The callback may freely mutate exported fields
// of g; no other Apply / Snapshot can interleave.
func (g *Game) WithWriteLock(fn func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fn()
}

// ReadSnapshot runs fn while holding a read lock on the game. The
// callback gets to read any exported field of g consistently — no
// other mutations can interleave. Callers MUST NOT mutate the game
// inside fn; use the exported mutation methods (which take a write
// lock) for that.
//
// S16: ReadSnapshot opportunistically resolves any stale layer-
// system recompute before invoking fn so the projection sees the
// current effective characteristics. The fast path — two atomic
// loads under the read lock — is what almost every broadcast hits,
// because every layer-version bump happens under the WRITE lock and
// most mutators recompute before they release it.
//
// When the cache IS stale the recompute cannot run here, because a
// recompute is a write: it reassigns Card.effective on every
// battlefield card, and since the S24 layer-2 control change it also
// writes Card.Controller (plus the CR 302.6 / 506.4 consequences) on
// any permanent whose controller just moved. A sync.RWMutex read
// lock is shared, so doing that under it races every other read-lock
// holder — `Game.AutoTapForCostExcluding` from the lobby's /autotap
// handler and `Game.ControllerOfCard` from the actions package both
// read Card.Controller off the battlefield on their own goroutines.
// So we drop the read lock, retake it in write mode for the
// recompute alone, and then re-enter read mode.
//
// The loop re-checks rather than assuming: a mutation can land in
// the gap between releasing the write lock and retaking the read
// lock, and fn() must see fresh layers. Each extra iteration costs
// one intervening mutation, so it terminates for the same reason the
// engine makes progress at all.
//
// This exists so that the protocol package can build a wire-format
// view of the game without the game package depending on the protocol
// package (which would be a circular import, since protocol views are
// built from game types).
func (g *Game) ReadSnapshot(fn func()) {
	g.mu.RLock()
	for g.layerVersion.Load() != g.lastResolvedVersion.Load() {
		g.mu.RUnlock()
		g.mu.Lock()
		g.RecomputeLayersIfStaleLocked()
		g.mu.Unlock()
		g.mu.RLock()
	}
	defer g.mu.RUnlock()
	fn()
}

// Snapshot returns a shallow copy of the top-level game state suitable
// for logging or diffing. Slices and maps are aliased, so a Snapshot
// is a read-only view and must not be mutated. For wire-format
// snapshots, use protocol.ViewOfGame instead.
func (g *Game) Snapshot() Game {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return Game{
		ID:          g.ID,
		CreatedAt:   g.CreatedAt,
		State:       g.State,
		Seats:       g.Seats,
		Battlefield: g.Battlefield,
		Stack:       g.Stack,
		Exile:       g.Exile,
		Turn:        g.Turn,
	}
}
