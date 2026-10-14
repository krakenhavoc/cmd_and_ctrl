package legal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// companion_test.go — the companion special action (CR 116.2g, ADR
// 0172 §7) is offered exactly when the engine takes it (#544).

// newCompanionTable is newTable with seat 0 revealing a Jegantha.
func newCompanionTable(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	var companionID uuid.UUID
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			c := game.NewCard("Jegantha, the Wellspring", uuid.Nil)
			c.OracleID = "f0587c55-06c1-4930-a62e-668f37464ce9"
			c.TypeLine = "Legendary Creature — Elemental Elk"
			c.ManaCost = "{4}{R/G}"
			c.IsCompanion = true
			companionID = c.InstanceID
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g, companionID
}

func TestCompanionMoveIsOfferedExactlyWhenTheEngineTakesIt(t *testing.T) {
	g, id := newCompanionTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("seat %d is active", g.Turn.ActiveSeat)
	}

	// The upkeep: not a main phase.
	if acts := specialActionsOf(legal.EnumerateFor(g, me.ID), id); len(acts) != 0 {
		t.Fatalf("offered in the upkeep: %v", labels(acts))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	// No mana: the {3} is unaffordable.
	if acts := specialActionsOf(legal.EnumerateFor(g, me.ID), id); len(acts) != 0 {
		t.Fatalf("offered with no mana: %v", labels(acts))
	}
	for range 3 {
		battlefieldCard(g, me, basic("Island", "Island"))
	}
	moves := legal.EnumerateFor(g, me.ID)
	acts := specialActionsOf(moves, id)
	if len(acts) != 1 {
		t.Fatalf("want one companion move, got %v", labels(acts))
	}
	if want := "Pay {3}: put Jegantha, the Wellspring into your hand (companion)"; acts[0].Label != want {
		t.Errorf("label %q, want %q", acts[0].Label, want)
	}
	dispatchAll(t, g, me.ID, acts)
	// Never offered to anyone else.
	if acts := specialActionsOf(legal.EnumerateFor(g, opp.ID), id); len(acts) != 0 {
		t.Errorf("offered to the opponent: %v", labels(acts))
	}
}
