package game

import (
	"math/rand/v2"
	"sync"
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
// internal mutex. Callers outside this package never touch fields
// directly — the exported methods are the whole surface.
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

	mu sync.Mutex
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

// Start transitions the game from lobby to active, initialises the
// turn cursor at seat 0 / turn 1 / untap step, and shuffles each
// player's library using the supplied RNG (nil for the package
// default). Returns ErrNotEnoughPlayers if fewer than MinPlayers are
// seated, and ErrGameAlreadyStarted if the game is not in lobby.
func (g *Game) Start(r *rand.Rand) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateLobby {
		return ErrGameAlreadyStarted
	}
	if len(g.Seats) < MinPlayers {
		return ErrNotEnoughPlayers
	}

	for _, p := range g.Seats {
		p.Library.Shuffle(r)
	}
	g.Turn = newStartingTurn()
	g.State = StateActive
	return nil
}

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
// Returns ErrGameNotActive if the game is not in the active state.
func (g *Game) AdvanceStep() (Turn, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return Turn{}, ErrGameNotActive
	}
	g.Turn = g.Turn.advance(len(g.Seats))
	return g.Turn, nil
}

// ActivePlayer returns the player whose turn it currently is, or nil
// if the game has not yet started.
func (g *Game) ActivePlayer() *Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive || len(g.Seats) == 0 {
		return nil
	}
	return g.Seats[g.Turn.ActiveSeat]
}

// PlayerByID looks up a seated player by ID. Returns nil if no player
// with that ID is seated.
func (g *Game) PlayerByID(id uuid.UUID) *Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Snapshot returns a shallow copy of the top-level game state suitable
// for logging or diffing. Slices and maps are aliased, so a Snapshot
// is a read-only view and must not be mutated. Use this for tests and
// diagnostics; the S03 protocol layer will build proper wire-format
// snapshots on top of this.
func (g *Game) Snapshot() Game {
	g.mu.Lock()
	defer g.mu.Unlock()
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
