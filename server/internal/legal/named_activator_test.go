package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// named_activator_test.go — the enumerator half of ADR 0106 §1's
// 2026-10-07 amendment (#1947): an opponents-only or owner-only row is a
// move for exactly the seats MayActivate names, and the dispatcher
// accepts every one of them (#544).

const (
	oracleClergy      = "66566999-f70a-4f14-9bf0-23325295a977"
	oracleIncarnation = "6e49a5b8-6bc4-4c7b-82c1-957f1fb0ca5f"
)

func TestOpponentsOnlyRowIsNotTheControllersMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	clergy := battlefieldCard(g, active, game.Card{
		Name: "Clergy of the Holy Nimbus", TypeLine: "Creature — Human Cleric", OracleID: oracleClergy, Power: 1, Toughness: 1,
	})
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), clergy); len(acts) != 0 {
		t.Fatalf("the controller was offered the opponents-only row: %v", labels(acts))
	}
}

func TestOpponentsOnlyRowIsAnOpponentsMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	clergy := battlefieldCard(g, controller, game.Card{
		Name: "Clergy of the Holy Nimbus", TypeLine: "Creature — Human Cleric", OracleID: oracleClergy, Power: 1, Toughness: 1,
	})
	mana(g, active, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), clergy)
	if len(acts) != 1 {
		t.Fatalf("want one activation for the opponent, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}

func TestOwnerOnlyRowIsTheOwnersMoveNotTheThiefs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	// Owned by the active seat, controlled by another (stolen).
	inc := battlefieldCard(g, active, game.Card{
		Name: "Personal Incarnation", TypeLine: "Creature — Avatar Incarnation", OracleID: oracleIncarnation, Power: 6, Toughness: 6,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == inc {
			g.Battlefield.Cards[i].Controller = other.ID
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), inc)
	if len(acts) != 1 {
		t.Fatalf("the owner of a stolen Incarnation: want one activation, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	// Back in its owner's hands, the controller is the owner again.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == inc {
			g.Battlefield.Cards[i].Controller = active.ID
			g.Battlefield.Cards[i].Owner = other.ID
		}
	}
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), inc); len(acts) != 0 {
		t.Fatalf("the thief was offered the owner-only row: %v", labels(acts))
	}
}
