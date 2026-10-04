package legal_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ADR 0121 §1: while the opening roll is open nobody has a hand, so the
// keep and mulligan moves the MulligansOpen branch would offer are not
// offered. The roll's own moves arrive with ADR 0121 PR 3 (§4).
func TestNoMovesWhileTheOpeningRollIsOpen(t *testing.T) {
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
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(1, 2027))); err != nil { // 20 17
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if moves := legal.EnumerateFor(g, p.ID); len(moves) != 0 {
			t.Fatalf("seat %d is offered %d moves during the opening roll: %+v", p.Seat, len(moves), moves)
		}
	}
	for _, p := range g.Seats {
		if err := g.RollOpening(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.ChooseStartingPlayer(g.Seats[0].ID, 0); err != nil {
		t.Fatal(err)
	}
	if moves := legal.EnumerateFor(g, g.Seats[0].ID); len(moves) == 0 {
		t.Fatal("no keep or mulligan move after the deal")
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
