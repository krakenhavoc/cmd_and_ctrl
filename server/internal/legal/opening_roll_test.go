package legal_test

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// opening_roll_test.go — ADR 0121 §4: while the opening roll is open a
// seat is offered its own die, the winner one choice per seat still in
// the game, and nobody anything else. Every offered move is accepted
// (dispatchAll, #499 / #544).

func openingRollTable(t *testing.T, n int, seed1, seed2 uint64) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range n {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(seed1, seed2))); err != nil {
		t.Fatal(err)
	}
	return g
}

// assertOnlyOpeningRollMoves fails on any move that is not one of the
// roll's two seat verbs — never keep or mulligan (no hand exists),
// never host_roll_remaining (a bot is never the host), never a table
// roll.
func assertOnlyOpeningRollMoves(t *testing.T, seat int, moves []legal.Move) {
	t.Helper()
	for _, m := range moves {
		if m.Kind != legal.KindOpeningRoll ||
			(m.Type != legal.TypeRollOpening && m.Type != legal.TypeChooseStartingPlayer) {
			t.Errorf("seat %d is offered %s (%s, %q) during the opening roll", seat, m.Type, m.Kind, m.Label)
		}
	}
}

func chosenSeat(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		Seat *int `json:"seat"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil || p.Seat == nil {
		t.Fatalf("choose_starting_player params %s: %v", m.Params, err)
	}
	return *p.Seat
}

// Key (1, 2027) at three seats rolls 20, 17, 1: seat 0 wins.
func TestOpeningRollOffersEachSeatItsDieThenTheWinnerTheChoice(t *testing.T) {
	g := openingRollTable(t, 3, 1, 2027)

	for _, p := range g.Seats {
		moves := legal.EnumerateFor(g, p.ID)
		assertOnlyOpeningRollMoves(t, p.Seat, moves)
		if len(moves) != 1 || moves[0].Type != legal.TypeRollOpening || !moves[0].AlwaysLegal || moves[0].Player != p.ID {
			t.Fatalf("seat %d before rolling: %+v, want one AlwaysLegal roll_opening", p.Seat, moves)
		}
		dispatchAll(t, g, p.ID, moves)
	}

	// A seat that has rolled has nothing to do while the others roll.
	if err := g.RollOpening(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if moves := legal.EnumerateFor(g, g.Seats[1].ID); len(moves) != 0 {
		t.Fatalf("seat 1 after rolling is offered %+v", moves)
	}
	for _, seat := range []int{0, 2} {
		if moves := legal.EnumerateFor(g, g.Seats[seat].ID); len(moves) != 1 || moves[0].Type != legal.TypeRollOpening {
			t.Fatalf("seat %d still owes its die, offered %+v", seat, moves)
		}
	}
	for _, seat := range []int{2, 0} {
		if err := g.RollOpening(g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	if g.OpeningRoll == nil || g.OpeningRoll.Chooser != 0 {
		t.Fatalf("opening roll after every die: %+v, want seat 0 choosing", g.OpeningRoll)
	}

	// The winner chooses among every seat, its own included, and only
	// its own is AlwaysLegal: another seat can concede before the
	// choice lands.
	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	assertOnlyOpeningRollMoves(t, 0, moves)
	if len(moves) != 3 {
		t.Fatalf("the chooser is offered %d moves, want one per seat: %+v", len(moves), moves)
	}
	for i, m := range moves {
		seat := chosenSeat(t, m)
		if seat != i || m.Type != legal.TypeChooseStartingPlayer || m.Player != g.Seats[0].ID {
			t.Fatalf("choice %d: %+v", i, m)
		}
		if m.AlwaysLegal != (seat == 0) {
			t.Errorf("choice of seat %d AlwaysLegal=%v, want only the chooser's own", seat, m.AlwaysLegal)
		}
	}
	if moves[0].Label != "I go first" || moves[1].Label != "P2 goes first" {
		t.Errorf("labels %q, %q", moves[0].Label, moves[1].Label)
	}
	dispatchAll(t, g, g.Seats[0].ID, moves)
	for _, seat := range []int{1, 2} {
		if moves := legal.EnumerateFor(g, g.Seats[seat].ID); len(moves) != 0 {
			t.Fatalf("seat %d is offered %+v while the winner chooses", seat, moves)
		}
	}

	// After the choice the window has closed and the mulligan is next.
	if err := g.ChooseStartingPlayer(g.Seats[0].ID, 0); err != nil {
		t.Fatal(err)
	}
	if moves := legal.EnumerateFor(g, g.Seats[0].ID); countKind(moves, legal.KindMulligan) == 0 || countKind(moves, legal.KindOpeningRoll) != 0 {
		t.Fatalf("after the deal: %+v, want keep and mulligan only", moves)
	}
}

// Key (22, 2048) at four seats: round 1 is 7, 4, 8, 8, so seats 2 and 3
// roll again; seats 0 and 1 are out of it.
func TestOpeningRollOffersARerollOnlyToTheTiedLeaders(t *testing.T) {
	g := openingRollTable(t, 4, 22, 2048)
	for _, p := range g.Seats {
		if err := g.RollOpening(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	if g.OpeningRoll == nil || len(g.OpeningRoll.Rounds) != 2 {
		t.Fatalf("after round 1: %+v, want a second round", g.OpeningRoll)
	}
	for _, p := range g.Seats {
		moves := legal.EnumerateFor(g, p.ID)
		assertOnlyOpeningRollMoves(t, p.Seat, moves)
		tied := p.Seat == 2 || p.Seat == 3
		if tied && (len(moves) != 1 || moves[0].Type != legal.TypeRollOpening) {
			t.Fatalf("tied seat %d offered %+v, want its reroll", p.Seat, moves)
		}
		if !tied && len(moves) != 0 {
			t.Fatalf("seat %d is out of round 2 and offered %+v", p.Seat, moves)
		}
		dispatchAll(t, g, p.ID, moves)
	}
}

// ADR 0121 §4: the enumerator never offers roll_table_die. A table roll
// changes nothing in the game, and a random-tier bot would roll in
// every window. Checked in the mulligan and at the first priority
// window, after a table roll has been made.
func TestNoTableRollMoveIsEverOffered(t *testing.T) {
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	check := func(when string) {
		t.Helper()
		for _, p := range g.Seats {
			for _, m := range legal.EnumerateFor(g, p.ID) {
				if m.Type == "roll_table_die" {
					t.Fatalf("%s: seat %d is offered a table roll: %+v", when, p.Seat, m)
				}
			}
		}
	}
	if _, err := g.RollTableDie(g.Seats[1].ID, game.TableDieD20); err != nil {
		t.Fatal(err)
	}
	check("the mulligan")
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	check("the first turn")
}

// A seat that leaves during the roll cannot be chosen.
func TestOpeningRollChooserCannotNameASeatThatLeft(t *testing.T) {
	g := openingRollTable(t, 3, 1, 2027) // 20 17 1
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeConcede, Player: g.Seats[2].ID, Caller: g.Seats[2].ID}); err != nil {
		t.Fatalf("concede during the roll: %v", err)
	}
	if moves := legal.EnumerateFor(g, g.Seats[2].ID); len(moves) != 0 {
		t.Fatalf("a seat that left is offered %+v", moves)
	}
	for _, seat := range []int{0, 1} {
		if err := g.RollOpening(g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	if len(moves) != 2 {
		t.Fatalf("the chooser is offered %+v, want seats 0 and 1", moves)
	}
	for _, m := range moves {
		if chosenSeat(t, m) == 2 {
			t.Fatalf("the chooser may name the seat that left: %+v", m)
		}
	}
	dispatchAll(t, g, g.Seats[0].ID, moves)
}
