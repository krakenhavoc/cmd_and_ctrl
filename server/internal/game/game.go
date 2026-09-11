package game

import (
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// rngSource is the minimal interface that zone.Shuffle needs from a
// randomness source. *math/rand/v2.Rand satisfies it; nil means
// "use the package-global source".
type rngSource = rand.Rand

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

	// Seats is the ordered list of players. Index matches Turn.ActiveSeat.
	Seats []*Player

	// Shared zones. Owner is uuid.Nil.
	Battlefield *Zone
	Stack       *Zone
	Exile       *Zone

	// Turn cursor, meaningful only when State == StateActive.
	Turn Turn

	// MulligansOpen is true between Start and the moment all seated
	// players have called KeepHand. While open, the client is
	// expected to surface a keep / mulligan dialog and gate the
	// "real" game UI behind the player's commitment. Added in S08.
	MulligansOpen bool

	// Monarch is the player ID currently designated as the monarch
	// (Conspiracy mechanic; player draws an extra card at end of their
	// turn). uuid.Nil means "no monarch currently". Set manually via
	// the set_monarch action — sandbox doesn't enforce the combat-damage
	// transfer rule. Added in S10.
	Monarch uuid.UUID

	// Initiative is the player ID currently designated as having taken
	// the initiative (Commander Legends: Battle for Baldur's Gate
	// mechanic; player ventures into the Undercity at the start of
	// their upkeep). uuid.Nil means "unassigned". Sandbox-only marker;
	// venturing is resolved manually until rules graft work lands.
	// Added in S10.
	Initiative uuid.UUID

	// UndoLimit is the per-player budget of undos allowed each turn.
	// Refreshed on each player's untap step. Default DefaultUndoLimit;
	// admin / any seated player can change via the set_undo_limit
	// action. Sandbox — players self-police, the limit is a guardrail
	// against runaway rewinds, not a strict policy. Added in S11.
	UndoLimit int

	// StartingSeat is the seat index that took the first turn. Set in
	// Start() to the active seat at game start. Used by the StepDraw
	// auto-action to skip the starting player's turn-1 draw per
	// CR 103.7c. Replays predating S13 default to seat 0 on decode,
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
	// SplitSecond set" (CR 702.79). While true, cast_spell and
	// activate_ability return ErrSplitSecondActive. Mana abilities
	// and special actions stay legal. Recomputed every time the
	// stack changes (cast, counter, resolve). Added in S13.1.
	SplitSecondActive bool

	// LoyaltyActivatedThisTurn flags planeswalkers whose loyalty
	// abilities have already been activated this turn (CR 606.5).
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

	// LandsPlayedThisTurn counts, per player, the lands that player
	// has played this turn via CastSpell's land branch. The engine
	// does NOT enforce the one-land-per-turn rule (CR 305.2) — that
	// stays a sandbox affordance — but the S31 legal-move enumerator
	// needs the count to offer a land drop only when one is still
	// owed. Cleared on Turn.advance to a new turn. Keyed by player
	// ID. Added in S31 sub-PR 1.
	LandsPlayedThisTurn map[uuid.UUID]int

	// DiscardPending is the cleanup-step pause map (S13.4): keys
	// are player IDs that need to discard, values are the count
	// each player must discard. Set at cleanup-step entry by
	// runStepEntryHooksLocked when Hand.Size() > MaxHandSize for
	// any non-eliminated player; cleared per-player by the
	// discard_selection action. The cleanup auto-advance is
	// blocked while this map is non-empty so the cursor pauses
	// for player input. Added in S13.4.
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

	// rng is captured from Start so that subsequent mutations that
	// shuffle (Mulligan, ShuffleLibrary) use the same source of
	// randomness as the initial library shuffle. nil means "use the
	// global math/rand/v2 source". Tests pass a deterministic
	// *rand.Rand here to get reproducible snapshots.
	rng *rngSource

	// rngState is the marshalable source backing rng, when the
	// engine minted it itself. Start(nil) — the only shape
	// production uses — seeds a *rand.PCG from crypto/rand and
	// keeps it here alongside the *rand.Rand that wraps it.
	//
	// This exists solely so the game's randomness can be
	// PERSISTED. math/rand/v2's *rand.Rand exposes no accessor for
	// its Source, so a game that only held `rng` could not have its
	// position in the random stream written to disk and resumed:
	// a restored game would shuffle from a different stream than
	// the one the players were in. *rand.PCG implements
	// encoding.BinaryMarshaler, so holding the source separately
	// makes the stream snapshot-able. See snapshot.go.
	//
	// nil when a caller supplied its own *rand.Rand (tests do, for
	// determinism) — that source's state is unreachable and the
	// snapshot records it as unpersistable rather than silently
	// reseeding. Also nil before Start.
	//
	// Not concurrency-safe on its own; every read goes through a
	// path already holding g.mu in write mode.
	rngState *rand.PCG

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

	// TurnScopedStatics is the per-turn CONTINUOUS-EFFECT slot —
	// the layer-engine twin of TurnScopedReplacements. Entries are
	// floating static abilities with a duration rather than a
	// battlefield source: Giant Growth's +3/+3, Overrun's mass pump
	// and trample grant, a loyalty ability's "+2/+2 and first strike
	// until end of turn". Consulted by activeStaticAbilitiesLocked
	// alongside the battlefield walk and swept at StepCleanup
	// (CR 514.2). See turn_scoped_statics.go. Added in S32.
	TurnScopedStatics []ScopedStatic

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
	replacementsAppliedThisEvent map[ReplacementEventID]map[ReplacementEffectID]bool

	// nextReplacementEventID mints the per-event keys stored in
	// replacementsAppliedThisEvent. Atomic so pipeline functions
	// can mint without promoting the lock (in practice they
	// already hold g.mu, but atomic is defensive). Added in S17
	// sub-PR 2.
	nextReplacementEventID atomic.Uint64

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
	}
	// S16 sub-PR 2: install the layer-engine invalidation listener.
	// Bumps g.layerVersion on the events that change which static
	// abilities are active (battlefield zone moves) or what they
	// apply to (counter changes). See layer_listener.go.
	g.Listeners = append(g.Listeners, layerVersionBump{})
	// S19 sub-PR 1: install the auto-fire trigger dispatcher. Walks
	// the battlefield (and the LKI map for LTB events) on every
	// emit, queues matching catalog-declared TriggeredAbility
	// entries onto PendingTriggers. See triggers.go.
	g.Listeners = append(g.Listeners, triggerHarvester{})
	// S17 sub-PR 2: install the CR 903.9 commander-zone built-in
	// replacement. Refactored from S13.1's inline
	// applyCommanderZoneReplacementLocked. See
	// builtin_replacements.go.
	g.BuiltinReplacements = append(g.BuiltinReplacements, commanderZoneReplacement)
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

	// Stamp every card with its owner (overwriting whatever the caller
	// set) and route commanders vs. library cards.
	for _, c := range deck {
		c.Owner = p.ID
		c.Controller = p.ID
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

	for _, c := range deck {
		c.Owner = p.ID
		c.Controller = p.ID
		if c.IsCommander {
			p.Command.PushTop(c)
		} else {
			p.Library.PushTop(c)
		}
	}
	p.DeckImported = true
	return nil
}

// Start transitions the game from lobby to active, initialises the
// turn cursor at seat 0 / turn 1 / untap step, and shuffles each
// player's library using the supplied RNG. The RNG is retained on
// the game and reused by subsequent shuffles (Mulligan,
// ShuffleLibrary) so that tests starting with a deterministic seed
// stay deterministic through all shuffles in the game — not just the
// initial one.
//
// Pass nil to have the engine mint its own crypto-seeded source.
// That is what production does, and unlike the process-global source
// it used to fall back to, an engine-owned source can be written to
// a snapshot and resumed after a restart (see Game.rngState).
//
// Returns ErrNotEnoughPlayers if fewer than MinPlayers are seated,
// and ErrGameAlreadyStarted if the game is not in lobby.
func (g *Game) Start(r *rand.Rand) error {
	// A caller-supplied source wins (tests seed deterministically).
	// Otherwise mint one the engine owns: a crypto-seeded *rand.PCG
	// wrapped in a *rand.Rand. Production always lands here.
	//
	// Before persistence this branch left g.rng nil and every
	// shuffle drew from math/rand/v2's process-global source. That
	// was unpredictable, but it was also unrecordable — a restart
	// could not resume the stream, it could only start a new one.
	// Owning the source makes the stream a piece of game state like
	// any other: it round-trips through a snapshot, so a deploy
	// resumes the library order the players were actually headed
	// for. The seed is minted from crypto/rand and never leaves the
	// server, so it stays unguessable.
	if r == nil {
		src := newCryptoSeededPCG()
		return g.start(rand.New(src), src)
	}
	// A *rand.Rand the caller built themselves: math/rand/v2 offers
	// no way to read its Source back out, so the stream position
	// cannot be snapshotted. The game runs fine; CaptureSnapshot
	// records it as unpersistable rather than silently reseeding on
	// restore. Use StartWithSource to get determinism AND
	// persistence.
	return g.start(r, nil)
}

// StartWithSource is Start on a caller-supplied PCG source.
//
// It exists because the two things a caller might want from Start's
// RNG argument — a reproducible stream, and a stream that survives a
// restart — are only compatible if the caller hands over the SOURCE
// rather than a *rand.Rand wrapping it. math/rand/v2's Rand keeps its
// Source private, so Start(rand.New(pcg)) throws away the only handle
// that could have been marshalled.
//
// Pass nil to get the same crypto-seeded source Start(nil) mints.
func (g *Game) StartWithSource(src *rand.PCG) error {
	if src == nil {
		src = newCryptoSeededPCG()
	}
	return g.start(rand.New(src), src)
}

// start is the shared body. rng is the source every shuffle in the
// game will draw from; state is its marshalable form, or nil when the
// caller owns a source we cannot read back.
func (g *Game) start(r *rand.Rand, state *rand.PCG) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return ErrGameAlreadyStarted
	}
	if len(g.Seats) < MinPlayers {
		return ErrNotEnoughPlayers
	}

	g.rng = r
	g.rngState = state
	if g.UndoLimit <= 0 {
		g.UndoLimit = DefaultUndoLimit
	}
	// S13.5: collect every seated player's ID so command-zone
	// initialisation can mark all knowers in one pass.
	allSeatedIDs := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		allSeatedIDs = append(allSeatedIDs, p.ID)
	}
	for _, p := range g.Seats {
		p.UndosRemaining = g.UndoLimit
		// g.rng, not r: when the caller passed nil we minted our own
		// source above, and the opening shuffle must draw from the
		// same stream every later shuffle will.
		p.Library.Shuffle(g.rng)
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
		// seated player sees them per CR 400.7e.
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(p.ID)
		}
		for i := range p.Command.Cards {
			p.Command.Cards[i].AddKnowersAll(allSeatedIDs)
		}
	}
	g.Turn = newStartingTurn()
	g.StartingSeat = g.Turn.ActiveSeat
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

// DefaultUndoLimit is the per-player undo budget refreshed each turn
// when Game.UndoLimit is unset. One is intentionally tight — undo is
// for "I clicked the wrong card", not for re-litigating turns. Admin
// or any seated player can raise it via set_undo_limit if the table
// wants more leniency. Added in S11.
const DefaultUndoLimit = 1

// End transitions the game to the ended state. Idempotent: calling
// End on an already-ended game is a no-op.
func (g *Game) End() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.State = StateEnded
}

// AdvanceStep moves the turn cursor forward by one step. After the
// cleanup step, the cursor wraps to the next seat's untap step and
// the turn number increments. Returns the new Turn.
//
// Side effects on step transitions:
//   - Entering combat_damage: ResolveCombatDamage runs (S08 —
//     auto-applies unblocked attacker damage to defending players'
//     life totals). Combat declarations (AttackingTarget /
//     BlockingTarget) are intentionally left in place.
//   - Entering end_combat: clearCombatLocked wipes the combat
//     declarations. Deferring the clear until this step lets the
//     client keep its combat-arrow overlay visible through the
//     entire combat_damage step instead of vanishing the moment
//     damage is resolved.
//
// Returns ErrGameNotActive if the game is not in the active state.
func (g *Game) AdvanceStep() (Turn, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return Turn{}, ErrGameNotActive
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

// advanceCursorLocked moves the step cursor forward by one — the
// single seam every step transition goes through. On a turn wrap it
// skips seats that have left the game (CR 800.4a: an eliminated
// player's turns are skipped) and fires onTurnAdvanceLocked so the
// per-turn caches clear.
//
// Both halves fix bugs the S31 bot fuzzer found on its first run:
// Turn.advance rotated into eliminated seats, handing priority to a
// player who could not act (humans had been escaping with
// advance_step), and the cleanup hook's wrap never called
// onTurnAdvanceLocked, so SpellsCastThisTurn / LoyaltyActivatedThisTurn
// survived every ordinary turn change. Caller must hold g.mu.
func (g *Game) advanceCursorLocked() {
	prev := g.Turn
	n := len(g.Seats)
	g.Turn = g.Turn.advance(n)
	if prev.IsNewTurn(g.Turn) {
		for i := 0; i < n && g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < n && g.Seats[g.Turn.ActiveSeat].Eliminated; i++ {
			// Wrap again from this seat's (never-taken) cleanup.
			skipped := g.Turn
			skipped.Step = StepCleanup
			g.Turn = skipped.advance(n)
		}
	}
	g.onTurnAdvanceLocked(prev, g.Turn)
}

// onTurnAdvanceLocked clears any per-turn caches whenever the
// active seat changes. Currently flushes
// `LoyaltyActivatedThisTurn` (CR 606.5 — once per turn per
// planeswalker), but the same hook is the natural home for any
// other "reset on new turn" caches the engine grows. Caller must
// hold g.mu.
func (g *Game) onTurnAdvanceLocked(prev, next Turn) {
	if !prev.IsNewTurn(next) {
		return
	}
	// S25 (#77): invalidate the layer cache. A continuous effect
	// whose AppliesTo reads the TURN rather than the battlefield —
	// Zurgo Helmsmasher's "during your turn, ~ has indestructible" is
	// the first in the catalog — changes its answer here and nowhere
	// else, so nothing would otherwise mark the cached
	// characteristics stale and the keyword would stick around on the
	// wrong player's turn.
	//
	// This is the bump layer_listener.go's header predicted and
	// deliberately deferred ("Step / phase advance … when they
	// arrive in a later sprint, advance the version inside the
	// step-advance helper directly — no event for it today"). It
	// lands here rather than in the listener for the reason that note
	// gives: there is no event for a turn change to listen to.
	//
	// Cost is one recompute per turn, against a cache that is already
	// invalidated by every zone move and every counter placed.
	g.layerVersion.Add(1)
	if g.LoyaltyActivatedThisTurn != nil {
		g.LoyaltyActivatedThisTurn = nil
	}
	if g.SpellsCastThisTurn != nil {
		g.SpellsCastThisTurn = nil
	}
	if g.LandsPlayedThisTurn != nil {
		g.LandsPlayedThisTurn = nil
	}
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
//   - StepUntap (S13): untap all permanents the active seat controls;
//     refresh their per-turn undo budget; auto-advance because Untap
//     grants no priority.
//   - StepDraw (S13): draw 1 for the active seat, except when the
//     starting player would draw on turn 1 (CR 103.7c skip).
//   - StepCombatDamage: auto-resolve unblocked attacker damage.
//   - StepEnd (S22): emit EventBeginEndStep so "at the beginning of
//     your end step" triggers fire. The delayed-trigger drain that
//     runs just before the switch covers "at the beginning of the
//     next end step" for the same boundary.
//   - StepEndCombat: clear AttackingTarget / BlockingTarget on every
//     battlefield card. Deferring the clear until end_combat (rather
//     than combat_damage) lets the client keep its combat-arrow
//     overlay visible through the entire combat_damage step instead
//     of vanishing the moment damage is resolved.
//   - StepCleanup (S13): auto-advance past cleanup to the next
//     seat's turn. S13.4 will hook interactive discard in here.
//
// Caller must hold g.mu.
func (g *Game) runStepEntryHooksLocked() {
	// S17 sub-PR 2: step-transition replacement hook. Stasis
	// cancels StepUntap; "skip your next upkeep" cards would
	// cancel StepUpkeep. The engine short-circuits on the cancel
	// path: advance to the next step and recurse through this
	// hook. Apply-loop iteration for skip-step is order-
	// independent (multiple "skip this step" effects are
	// idempotent) so even with ≥2 applicable the prompt path
	// never actually queues — gatherActiveReplacementsLocked
	// returns at most one eligible replacement in practice for
	// sub-PR 2 (zero catalog replacements registered).
	stepEv := &ReplacementEvent{
		Kind:               RepEventStepTransition,
		StepTransitionStep: g.Turn.Step,
		StepTransitionSeat: g.Turn.ActiveSeat,
	}
	out, err := g.applyReplacementsLocked(stepEv)
	if !errors.Is(err, errReplacementPending) {
		defer g.clearReplacementEventLocked(stepEv.ID)
		// Canceled events come back as (nil, nil) from
		// applyReplacementsLocked — check err==nil + out==nil as
		// the cancel signal, plus the belt-and-braces out.Canceled
		// for any intermediate path that returns the event.
		canceled := err == nil && (out == nil || out.Canceled)
		if canceled {
			// Step canceled — advance past and recurse so the
			// cursor hits the next step's entry hook.
			g.advanceCursorLocked()
			g.runStepEntryHooksLocked()
			return
		}
	}
	// CR 106.4: every player's mana pool empties at the end of each
	// step / phase. We model this by clearing pools at the START of
	// the next step's entry — equivalent net effect, and centralised
	// here so every step transition (including the auto-advance
	// recursion through Untap → Upkeep and Cleanup → next-Untap)
	// triggers the clear. Cheap to call on already-empty pools.
	g.emptyAllManaPoolsLocked()
	// S22: CR 603.7 delayed triggered abilities fire on ENTRY to the
	// step they name. Draining here — before the per-step turn-based
	// actions below — is what makes "the NEXT end step" work without
	// any created-this-step bookkeeping: an ability scheduled during
	// an end step is queued after this hook has already run for that
	// step, so it waits for the following one. See delayed.go.
	g.fireDelayedTriggersLocked(g.Turn.Step)
	switch g.Turn.Step {
	case StepPrecombatMain:
		// S27 / CR 714.2b: "after your draw step, put a lore counter
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
		// reason CR 714.2b puts that first: the turn-based action
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
		// fall through. Mulligans keep the cursor parked at Untap, so
		// this is normally unreachable while they're open; guard
		// anyway.
		if g.MulligansOpen {
			return
		}
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			g.EmitEvent(Event{
				Kind:  EventBeginUpkeep,
				Actor: g.Seats[g.Turn.ActiveSeat].ID,
			})
		}
	case StepUntap:
		// Mulligans still open → hold the cursor at Untap until
		// KeepHand closes the window. KeepHand re-runs this hook.
		if g.MulligansOpen {
			return
		}
		if g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
			g.Seats[g.Turn.ActiveSeat].UndosRemaining = g.UndoLimit
			g.untapAllForLocked(g.Turn.ActiveSeat)
		}
		// Untap grants no priority; recurse into the next step.
		g.advanceCursorLocked()
		g.runStepEntryHooksLocked()
	case StepDraw:
		if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
			return
		}
		// CR 103.7c: the player who takes the first turn skips their
		// draw step on turn 1. Subsequent turns are normal.
		if g.Turn.Number == 1 && g.Turn.ActiveSeat == g.StartingSeat {
			return
		}
		// Best-effort: an empty library on auto-draw is not a hard
		// error here (losing from drawing from an empty library is a
		// SBA that lands in S13.1). The drawCardLocked helper surfaces
		// ErrZoneEmpty, which we swallow so the cursor keeps moving —
		// the losing player will be caught by the SBA once it exists.
		_ = g.drawCardLocked(g.Seats[g.Turn.ActiveSeat].ID)
	case StepCombatDamage:
		g.resolveCombatDamageLocked()
	case StepEndCombat:
		g.clearCombatLocked()
	case StepCleanup:
		// CR 402.2: build the discard-pending map for any player
		// over their per-player MaxHandSize. The cursor pauses at
		// cleanup until every entry is drained via discard_selection
		// (S13.4 — the interactive discard pathway).
		g.populateDiscardPendingLocked()
		// CR 514.2: damage marked on permanents is removed at the
		// start of cleanup, regardless of whether the discard pause
		// fires. The lethal-damage SBA from S13.1 reads DamageMarked,
		// so clearing it here means the per-turn damage doesn't
		// carry over into the next turn.
		//
		// S18 sub-PR 3: MarkedLethalByDeathtouch is the companion
		// flag (CR 702.2c) set by combat damage from deathtouch
		// sources. Same per-turn scope as DamageMarked, cleared at
		// the same site.
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].DamageMarked = 0
			g.Battlefield.Cards[i].MarkedLethalByDeathtouch = false
		}
		// S17 sub-PR 5: "until end of turn" replacement effects
		// (Fog's prevent-all-combat-damage, future prevention
		// shields with a per-turn duration) clear at cleanup so
		// next turn starts with a clean slate.
		g.ClearTurnScopedReplacementsLocked()
		// S32: "until end of turn" CONTINUOUS effects (Giant
		// Growth's +3/+3, Overrun's trample grant) expire here for
		// the same reason and by the same rule — CR 514.2 ends them
		// during the cleanup step, before the turn-based discard.
		// This is also what makes a grant created during the END
		// step end this turn rather than next: the sweep keys on the
		// turn number stamped at registration, not on "the next
		// cleanup after the one I saw".
		g.ClearExpiredTurnScopedStaticsLocked()
		// S21 sub-PR 6: impulse-exile permissions ("you may play it
		// this turn") lapse here for the same reason — the turn they
		// were granted for is over. The exiled card stays exiled; it
		// just stops being playable.
		g.clearExpiredExilePlayLocked()
		// Auto-advance only when no player owes discard. Otherwise
		// the cursor sits at Cleanup with PriorityHolder=NoPriority
		// until DiscardSelection drains the pending map and re-fires
		// this hook.
		if len(g.DiscardPending) == 0 {
			g.advanceCursorLocked()
			g.runStepEntryHooksLocked()
		}
	}
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
