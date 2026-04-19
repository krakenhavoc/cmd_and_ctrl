package game

import (
	"math/rand/v2"
	"sync"
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

	// rng is captured from Start so that subsequent mutations that
	// shuffle (Mulligan, ShuffleLibrary) use the same source of
	// randomness as the initial library shuffle. nil means "use the
	// global math/rand/v2 source". Tests pass a deterministic
	// *rand.Rand here to get reproducible snapshots.
	rng *rngSource

	mu sync.RWMutex
}

// NewGame constructs a game in the lobby state with a fresh ID and
// empty shared zones.
func NewGame() *Game {
	return &Game{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		State:       StateLobby,
		Battlefield: newZone(ZoneBattlefield, uuid.Nil),
		Stack:       newZone(ZoneStack, uuid.Nil),
		Exile:       newZone(ZoneExile, uuid.Nil),
	}
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
	for _, p := range g.Seats {
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
	}
	g.Turn = newStartingTurn()
	g.State = StateActive
	g.MulligansOpen = true
	return nil
}

// OpeningHandSize is the number of cards each player draws when the
// game starts. The mulligan flow can take a player back to the same
// count for redraws (simplified — no London bottom-N penalty yet).
const OpeningHandSize = 7

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
	g.Turn = g.Turn.advance(len(g.Seats))
	switch g.Turn.Step {
	case StepCombatDamage:
		g.resolveCombatDamageLocked()
	case StepEndCombat:
		g.clearCombatLocked()
	}
	return g.Turn, nil
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

// ReadSnapshot runs fn while holding a read lock on the game. The
// callback gets to read any exported field of g consistently — no
// other mutations can interleave. Callers MUST NOT mutate the game
// inside fn; use the exported mutation methods (which take a write
// lock) for that.
//
// This exists so that the protocol package can build a wire-format
// view of the game without the game package depending on the protocol
// package (which would be a circular import, since protocol views are
// built from game types).
func (g *Game) ReadSnapshot(fn func()) {
	g.mu.RLock()
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
