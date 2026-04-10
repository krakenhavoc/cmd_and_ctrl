package game

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// buildTestDeck returns a synthetic 99-card deck plus a commander,
// owned by a placeholder UUID (the real owner is stamped by AddPlayer).
// Card identities and names are placeholders — S02 does not care
// about real card data, which arrives with the Scryfall pipeline in
// S04.
func buildTestDeck(commanderName string) []Card {
	owner := uuid.Nil
	deck := make([]Card, 0, 100)
	deck = append(deck, NewCommander(commanderName, owner))
	for i := 0; i < 99; i++ {
		deck = append(deck, NewCard(fmt.Sprintf("Filler Card %d", i+1), owner))
	}
	return deck
}

func TestNewGameStartsInLobby(t *testing.T) {
	g := NewGame()
	if g.State != StateLobby {
		t.Errorf("state: got %q, want %q", g.State, StateLobby)
	}
	if g.ID == uuid.Nil {
		t.Error("game ID is nil")
	}
	if g.Battlefield == nil || g.Stack == nil || g.Exile == nil {
		t.Error("shared zones should be pre-constructed")
	}
	if g.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestAddPlayerRoutesCommanderToCommandZone(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("Alice", buildTestDeck("Atraxa, Praetors' Voice"))
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if p.Command.Size() != 1 {
		t.Errorf("command zone: got %d cards, want 1", p.Command.Size())
	}
	if p.Library.Size() != 99 {
		t.Errorf("library: got %d cards, want 99", p.Library.Size())
	}
	cmdr, _ := p.Command.Top()
	if !cmdr.IsCommander {
		t.Error("card in command zone should be flagged as commander")
	}
	if cmdr.Owner != p.ID {
		t.Error("commander owner should be stamped to player ID by AddPlayer")
	}
}

func TestAddPlayerRejectsEmptyName(t *testing.T) {
	g := NewGame()
	if _, err := g.AddPlayer("", buildTestDeck("X")); err != ErrEmptyName {
		t.Errorf("empty name: got %v, want ErrEmptyName", err)
	}
}

func TestAddPlayerRejectsEmptyDeck(t *testing.T) {
	g := NewGame()
	if _, err := g.AddPlayer("A", nil); err != ErrEmptyDeck {
		t.Errorf("empty deck: got %v, want ErrEmptyDeck", err)
	}
}

func TestAddPlayerFullGame(t *testing.T) {
	g := NewGame()
	for i := 0; i < MaxPlayers; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C")); err != nil {
			t.Fatalf("AddPlayer #%d: %v", i+1, err)
		}
	}
	if _, err := g.AddPlayer("Overflow", buildTestDeck("C")); err != ErrGameFull {
		t.Errorf("5th player: got %v, want ErrGameFull", err)
	}
}

func TestStartRequiresMinPlayers(t *testing.T) {
	g := NewGame()
	if err := g.Start(nil); err != ErrNotEnoughPlayers {
		t.Errorf("start with 0 players: got %v, want ErrNotEnoughPlayers", err)
	}
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	if err := g.Start(nil); err != ErrNotEnoughPlayers {
		t.Errorf("start with 1 player: got %v, want ErrNotEnoughPlayers", err)
	}
}

func TestStartTransitionsAndShuffles(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}

	// Capture pre-shuffle order to verify Start actually shuffles.
	before := make([]uuid.UUID, len(g.Seats[0].Library.Cards))
	for i, c := range g.Seats[0].Library.Cards {
		before[i] = c.InstanceID
	}

	// Use a deterministic RNG so this test is reproducible.
	r := rand.New(rand.NewPCG(1, 2))
	if err := g.Start(r); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.State != StateActive {
		t.Errorf("state: got %q, want %q", g.State, StateActive)
	}
	if g.Turn.Number != 1 || g.Turn.ActiveSeat != 0 || g.Turn.Step != StepUntap {
		t.Errorf("starting turn: got %+v", g.Turn)
	}

	// Verify the library was actually shuffled — at least one index
	// should differ after shuffling a 99-card list (probability of a
	// no-op shuffle is astronomically small).
	differs := false
	for i, c := range g.Seats[0].Library.Cards {
		if c.InstanceID != before[i] {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("library order is unchanged after Start; shuffle did not run")
	}
}

func TestDoubleStartRejected(t *testing.T) {
	g := NewGame()
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	_, _ = g.AddPlayer("B", buildTestDeck("C"))
	_ = g.Start(nil)
	if err := g.Start(nil); err != ErrGameAlreadyStarted {
		t.Errorf("double start: got %v, want ErrGameAlreadyStarted", err)
	}
}

func TestAdvanceStepRequiresActiveGame(t *testing.T) {
	g := NewGame()
	if _, err := g.AdvanceStep(); err != ErrGameNotActive {
		t.Errorf("advance in lobby: got %v, want ErrGameNotActive", err)
	}
}

// TestFullFourPlayerTurnCycle is the S02 exit criteria test: it
// stands up a 4-player Commander game, steps through a complete turn
// cycle, and asserts the expected state at each step along the way.
// If this passes, the domain types and turn state machine are sound
// enough to build the action protocol on top of in S03.
func TestFullFourPlayerTurnCycle(t *testing.T) {
	g := NewGame()
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	for i, name := range names {
		p, err := g.AddPlayer(name, buildTestDeck(fmt.Sprintf("Commander %d", i+1)))
		if err != nil {
			t.Fatalf("AddPlayer %s: %v", name, err)
		}
		if p.Seat != i {
			t.Errorf("seat for %s: got %d, want %d", name, p.Seat, i)
		}
		if p.Life != StartingLife {
			t.Errorf("starting life for %s: got %d, want %d", name, p.Life, StartingLife)
		}
		if p.Command.Size() != 1 {
			t.Errorf("%s command zone: got %d, want 1", name, p.Command.Size())
		}
		if p.Library.Size() != 99 {
			t.Errorf("%s library: got %d, want 99", name, p.Library.Size())
		}
	}

	if len(g.Seats) != 4 {
		t.Fatalf("seats: got %d, want 4", len(g.Seats))
	}

	if err := g.Start(rand.New(rand.NewPCG(7, 13))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.State != StateActive {
		t.Fatalf("state after start: %q", g.State)
	}

	// The exit-criteria assertion: walk every step of seat 0's turn
	// and verify the cursor matches the canonical sequence.
	want := TurnSequence()
	active := g.ActivePlayer()
	if active == nil || active.Name != "Alice" {
		t.Fatalf("active player: got %v, want Alice", active)
	}

	// Cursor starts at index 0 (StepUntap) — verified directly.
	if g.Turn.Step != want[0] {
		t.Errorf("turn starts at step %q, want %q", g.Turn.Step, want[0])
	}
	if g.Turn.Phase != PhaseOf(want[0]) {
		t.Errorf("turn starts at phase %q, want %q", g.Turn.Phase, PhaseOf(want[0]))
	}

	// Walk steps 1..len-1 (the remaining 11 steps of Alice's turn).
	for i := 1; i < len(want); i++ {
		turn, err := g.AdvanceStep()
		if err != nil {
			t.Fatalf("AdvanceStep %d: %v", i, err)
		}
		if turn.Step != want[i] {
			t.Errorf("step %d: got %q, want %q", i, turn.Step, want[i])
		}
		if turn.Phase != PhaseOf(want[i]) {
			t.Errorf("step %d phase: got %q, want %q", i, turn.Phase, PhaseOf(want[i]))
		}
		if turn.ActiveSeat != 0 {
			t.Errorf("seat drifted off 0 mid-turn at step %d: seat=%d", i, turn.ActiveSeat)
		}
		if turn.Number != 1 {
			t.Errorf("turn number drifted off 1 mid-turn at step %d: number=%d", i, turn.Number)
		}
	}

	// One more AdvanceStep should wrap to seat 1, turn still 1, back
	// at untap (the beginning of Bob's first turn).
	turn, err := g.AdvanceStep()
	if err != nil {
		t.Fatalf("wrap AdvanceStep: %v", err)
	}
	if turn.ActiveSeat != 1 {
		t.Errorf("seat after wrap: got %d, want 1", turn.ActiveSeat)
	}
	if turn.Number != 1 {
		t.Errorf("turn number after 1→2 seat wrap: got %d, want 1", turn.Number)
	}
	if turn.Step != StepUntap {
		t.Errorf("step after wrap: got %q, want %q", turn.Step, StepUntap)
	}

	active = g.ActivePlayer()
	if active == nil || active.Name != "Bob" {
		t.Fatalf("active player after wrap: got %v, want Bob", active)
	}

	// Burn through seats 1, 2, 3 completely to verify full 4-player
	// round cycles correctly back to seat 0 with turn number 2.
	for seatsToFinish := 3; seatsToFinish > 0; seatsToFinish-- {
		// Each seat has 11 more steps to go after untap (we're already
		// at untap from the wrap/previous seat).
		for i := 1; i < len(want); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep during seat burn: %v", err)
			}
		}
		// Now advance once more to wrap to the next seat.
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep wrap: %v", err)
		}
	}

	if g.Turn.Number != 2 {
		t.Errorf("turn number after full round: got %d, want 2", g.Turn.Number)
	}
	if g.Turn.ActiveSeat != 0 {
		t.Errorf("seat after full round: got %d, want 0", g.Turn.ActiveSeat)
	}
	if g.Turn.Step != StepUntap {
		t.Errorf("step after full round: got %q, want %q", g.Turn.Step, StepUntap)
	}
}

func TestPlayerByID(t *testing.T) {
	g := NewGame()
	p, _ := g.AddPlayer("A", buildTestDeck("C"))
	if got := g.PlayerByID(p.ID); got != p {
		t.Errorf("PlayerByID: got %v, want %v", got, p)
	}
	if got := g.PlayerByID(uuid.New()); got != nil {
		t.Errorf("PlayerByID(unknown): got %v, want nil", got)
	}
}

func TestEndTransitionsToEnded(t *testing.T) {
	g := NewGame()
	_, _ = g.AddPlayer("A", buildTestDeck("C"))
	_, _ = g.AddPlayer("B", buildTestDeck("C"))
	_ = g.Start(nil)
	g.End()
	if g.State != StateEnded {
		t.Errorf("state after End: got %q, want %q", g.State, StateEnded)
	}
}
