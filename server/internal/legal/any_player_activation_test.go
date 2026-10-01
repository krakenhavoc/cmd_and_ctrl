package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_player_activation_test.go — the enumerator half of ADR 0106 §1
// (#1793). An "Any player may activate this ability" row on another
// player's permanent is a move for every seat with priority and the
// mana to pay, and the move is one the dispatcher accepts (#544). A
// row without the permission is never offered to a non-controller.

const oracleXantcha = "0f0f3712-8d13-41a5-b332-2ab34e48d79d"

func TestAnyPlayerRowIsOfferedToANonController(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	controller := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	clearHand(active)
	xantcha := battlefieldCard(g, owner, game.Card{
		Name: "Xantcha, Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		OracleID: oracleXantcha, Power: 5, Toughness: 5,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == xantcha {
			g.Battlefield.Cards[i].Controller = controller.ID
		}
	}
	// A permanent of the same controller whose ability is NOT anyone's.
	bombardment := battlefieldCard(g, controller, game.Card{Name: "Goblin Bombardment", TypeLine: "Enchantment", OracleID: oracleGoblinBombardment})
	battlefieldCard(g, controller, creature("Fodder", "{1}", 1, 1))
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, xantcha)
	if len(acts) != 1 {
		t.Fatalf("active non-controller: want one Xantcha activation, got %v", labels(moves))
	}
	if want := "Xantcha, Sleeper Agent (controlled by " + controller.Name + "): "; !strings.HasPrefix(acts[0].Label, want) {
		t.Errorf("label %q, want it to name the controller (%q…)", acts[0].Label, want)
	}
	if got := activationsOf(moves, bombardment); len(got) != 0 {
		t.Errorf("another player's controller-only ability offered: %v", labels(got))
	}
	dispatchAll(t, g, active.ID, acts)
}

// TestAnyPlayerRowNeedsTheActivatorsMana: the controller's lands do not
// make the activation affordable for anyone else (CR 602.1a).
func TestAnyPlayerRowNeedsTheActivatorsMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	xantcha := battlefieldCard(g, controller, game.Card{
		Name: "Xantcha, Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion", OracleID: oracleXantcha,
	})
	mana(g, controller, 3)
	mana(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), xantcha); len(acts) != 0 {
		t.Errorf("two mana of its own: want no activation, got %v", labels(acts))
	}
}
