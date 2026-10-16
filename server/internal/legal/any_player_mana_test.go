package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_player_mana_test.go — the enumerator half of #1864 (ADR 0106 §1
// amendment 2026-10-16): Mana Cache's "Remove a charge counter from this
// enchantment: Add {C}. Any player may activate this ability but only
// during their turn before the end step" is a mana move for the active
// player on whoever's Mana Cache it is, and for nobody else, and the move
// is one the dispatcher accepts (#544).

const oracleManaCache = "1c09f2d1-413b-4a92-aa39-99c1a2106814"

func manaCacheFor(g *game.Game, p *game.Player, charge int) uuid.UUID {
	id := battlefieldCard(g, p, game.Card{Name: "Mana Cache", TypeLine: "Enchantment", OracleID: oracleManaCache})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id && charge > 0 {
			g.Battlefield.Cards[i].Counters = map[string]int{game.CounterCharge: charge}
		}
	}
	return id
}

func TestAnyPlayerManaRowIsOfferedToTheActivePlayer(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	cache := manaCacheFor(g, controller, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	got := movesFrom(legal.EnumerateFor(g, active.ID), cache, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("active non-controller: want one Mana Cache mana move, got %v", labels(got))
	}
	if want := "Mana Cache (controlled by " + controller.Name + "): "; !strings.HasPrefix(got[0].Label, want) {
		t.Errorf("label %q, want it to name the controller (%q…)", got[0].Label, want)
	}
	if got[0].Player != active.ID {
		t.Errorf("move player = %v, want the activator %v", got[0].Player, active.ID)
	}
	dispatchAll(t, g, active.ID, got)

	// "Only during their turn": the controller and a bystander, on the
	// active player's turn, are offered nothing.
	for _, p := range []*game.Player{controller, bystander} {
		if moves := movesFrom(legal.EnumerateFor(g, p.ID), cache, legal.KindMana); len(moves) != 0 {
			t.Errorf("%s on another player's turn: offered %v", p.Name, labels(moves))
		}
	}
}

// TestAnyPlayerManaRowNeedsACounter: with no charge counter on it, the
// cost can't be paid and nobody is offered the row.
func TestAnyPlayerManaRowNeedsACounter(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cache := manaCacheFor(g, controller, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	if moves := movesFrom(legal.EnumerateFor(g, active.ID), cache, legal.KindMana); len(moves) != 0 {
		t.Errorf("no charge counter: offered %v", labels(moves))
	}
}

// TestAnyPlayerManaRowIsNotOfferedInTheEndStep: "before the end step".
func TestAnyPlayerManaRowIsNotOfferedInTheEndStep(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cache := manaCacheFor(g, controller, 2)
	advanceTo(t, g, game.StepEnd)
	if moves := movesFrom(legal.EnumerateFor(g, active.ID), cache, legal.KindMana); len(moves) != 0 {
		t.Errorf("end step: offered %v", labels(moves))
	}
}
