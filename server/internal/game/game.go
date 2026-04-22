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
	// production listeners — S19 registers the first real one.
	// Shallow-copied on Clone / RestoreFrom (listeners are process-
	// lifetime singletons). Added in S14 sub-PR 1.
	Listeners []Listener

	// rng is captured from Start so that subsequent mutations that
	// shuffle (Mulligan, ShuffleLibrary) use the same source of
	// randomness as the initial library shuffle. nil means "use the
	// global math/rand/v2 source". Tests pass a deterministic
	// *rand.Rand here to get reproducible snapshots.
	rng *rngSource

	// layerVersion is the S16 continuous-effect-engine invalidation
	// counter. Bumped by listeners on every event that could change
	// which static abilities are active or what they apply to —
	// battlefield zone changes, control changes, counter changes,
	// step advance. Sub-PR 1 ships the field; sub-PR 2 wires the
	// listener bumps; sub-PR 3 wires actual recompute work.
	// lastResolvedVersion mirrors the most recent version the
	// recompute pass has caught up to; the snapshot path no-ops when
	// they're equal. Both atomic so the snapshot path can call
	// RecomputeLayersIfStaleLocked under the read lock without a
	// data race against listener bumps. Read via .Load(), bumped via
	// .Add(1) / .Store(). Added in S16 sub-PR 1.
	layerVersion        atomic.Uint64
	lastResolvedVersion atomic.Uint64

	// recomputeCount instruments the layer engine for the fast-path
	// regression test (exit criterion #10): two consecutive snapshot
	// reads with no events between invoke RecomputeLayersLocked
	// exactly once. Test-only; production callers ignore. Atomic
	// for the same reason as the version counters.
	recomputeCount atomic.Uint64

	// recompute serialises layer recomputes against each other
	// without promoting the game's read lock. See layers.go.
	recompute recomputeState

	// BuiltinReplacements is the S17 replacement-effect registry
	// populated at NewGame. Today: just the commander-zone
	// replacement (refactored from S13.1's inline
	// applyCommanderZoneReplacementLocked). Future built-ins
	// register by appending here before Start. Catalog replacements
	// flow through CatalogReplacements (see effect_hooks.go) and
	// are NOT listed here. See replacements.go +
	// builtin_replacements.go. Added in S17 sub-PR 2.
	BuiltinReplacements []ReplacementEffect

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
// player's library using the supplied RNG (nil for the package
// default). The RNG is retained on the game and reused by subsequent
// shuffles (Mulligan, ShuffleLibrary) so that tests starting with a
// deterministic seed stay deterministic through all shuffles in the
// game — not just the initial one.
//
// Returns ErrNotEnoughPlayers if fewer than MinPlayers are seated,
// and ErrGameAlreadyStarted if the game is not in lobby.
func (g *Game) Start(r *rand.Rand) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return ErrGameAlreadyStarted
	}
	if len(g.Seats) < MinPlayers {
		return ErrNotEnoughPlayers
	}

	g.rng = r
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
		p.Library.Shuffle(r)
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
	prev := g.Turn
	g.Turn = g.Turn.advance(len(g.Seats))
	g.onTurnAdvanceLocked(prev, g.Turn)
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
	if g.LoyaltyActivatedThisTurn != nil {
		g.LoyaltyActivatedThisTurn = nil
	}
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
		if err == nil && out != nil && out.Canceled {
			// Step canceled — advance past and recurse so the
			// cursor hits the next step's entry hook.
			g.Turn = g.Turn.advance(len(g.Seats))
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
	switch g.Turn.Step {
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
		g.Turn = g.Turn.advance(len(g.Seats))
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
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].DamageMarked = 0
		}
		// Auto-advance only when no player owes discard. Otherwise
		// the cursor sits at Cleanup with PriorityHolder=NoPriority
		// until DiscardSelection drains the pending map and re-fires
		// this hook.
		if len(g.DiscardPending) == 0 {
			g.Turn = g.Turn.advance(len(g.Seats))
			g.runStepEntryHooksLocked()
		}
	}
}

// populateDiscardPendingLocked scans seated, non-eliminated players
// and records the over-max count for each one whose hand exceeds
// their MaxHandSize. NoMaxHandSize (-1) is treated as "no cap" and
// skipped. Caller must hold g.mu.
func (g *Game) populateDiscardPendingLocked() {
	g.DiscardPending = nil
	for _, p := range g.Seats {
		if p.Eliminated {
			continue
		}
		if p.MaxHandSize == NoMaxHandSize {
			continue
		}
		over := p.Hand.Size() - p.MaxHandSize
		if over <= 0 {
			continue
		}
		if g.DiscardPending == nil {
			g.DiscardPending = make(map[uuid.UUID]int)
		}
		g.DiscardPending[p.ID] = over
	}
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
// current effective characteristics. The fast-path no-ops when the
// version counters match; the body uses atomic version reads + a
// dedicated recompute mutex, so calling under the read lock stays
// race-free without promoting to the write lock.
//
// This exists so that the protocol package can build a wire-format
// view of the game without the game package depending on the protocol
// package (which would be a circular import, since protocol views are
// built from game types).
func (g *Game) ReadSnapshot(fn func()) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.RecomputeLayersIfStaleLocked()
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
