package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #2016: the discard payment the enumerator offers is the seat's
// cheapest N by Options.OrderCostFuel, not the first N in hand order.
func TestDiscardCostIsPaidWithTheCheapestCardsFirst(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	shaman := battlefieldCard(g, active, game.Card{
		Name: "Fauna Shaman", TypeLine: "Creature — Elf Shaman", Power: 2, Toughness: 2,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Discard two cards: do nothing",
			Cost:   game.AbilityCost{DiscardCards: &game.DiscardCost{N: 2, Label: "two cards"}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	dear1 := handCard(active, creature("Dragon", "{4}{R}{R}", 5, 5))
	dear2 := handCard(active, creature("Angel", "{3}{W}{W}", 4, 4))
	cheap1 := handCard(active, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	cheap2 := handCard(active, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), shaman)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	if ids := discardParamsOf(t, acts[0]).DiscardIDs; len(ids) != 2 || ids[0] != dear1.String() || ids[1] != dear2.String() {
		t.Fatalf("default order: discard_ids = %v, want hand order", ids)
	}

	price := func(c legal.TargetCandidate) float64 {
		if c.ID == cheap1 || c.ID == cheap2 {
			return 1
		}
		return 10
	}
	moves := legal.EnumerateForWithOptions(g, active.ID, legal.Options{OrderCostFuel: price})
	acts = activationsOf(moves, shaman)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	ids := discardParamsOf(t, acts[0]).DiscardIDs
	if len(ids) != 2 || ids[0] != cheap1.String() || ids[1] != cheap2.String() {
		t.Errorf("discard_ids = %v, want the two cheap cards %v %v", ids, cheap1, cheap2)
	}
	dispatchAll(t, g, active.ID, moves)
}
