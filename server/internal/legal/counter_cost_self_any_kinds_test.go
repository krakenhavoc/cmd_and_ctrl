package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_cost_self_any_kinds_test.go — the enumerator half of #3987:
// "Remove two counters from Ghost-Spider", two counters of any kinds off
// the source. The #544 invariant: offer exactly the payment the engine
// accepts (dispatchAll sends it through the real dispatcher), nothing when
// the source cannot pay, and never another permanent's counters.

func selfAnyKindsCost() game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{N: 2, Among: true}}
}

func TestSelfAnyKindsCounterCostEnumeratesAMixedPayment(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	src := counterCostCard("Self Any Kinds Test", selfAnyKindsCost())
	src.Counters = map[string]int{"+1/+1": 1}
	spider := battlefieldCard(g, active, src)
	// Another permanent's counters never pay.
	battlefieldCard(g, active, game.Card{
		Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{"+1/+1": 4},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), spider); len(acts) != 0 {
		t.Fatalf("one counter on the source against a printed two: want no activation, got %v", labels(acts))
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == spider {
				g.Battlefield.Cards[i].Counters[game.CounterShield] = 1
			}
		}
	})
	moves := activationsOf(legal.EnumerateFor(g, active.ID), spider)
	if len(moves) != 1 {
		t.Fatalf("want exactly one payment, got %v", labels(moves))
	}
	p := anyKindParamsOf(t, moves[0])
	if len(p.CounterSourceIDs) != 0 {
		t.Errorf("counter_source_ids = %v, want none (the self form names no permanent)", p.CounterSourceIDs)
	}
	if len(p.CounterCounts) != 2 || p.CounterCounts[0]+p.CounterCounts[1] != 2 {
		t.Errorf("counter_counts = %v, want two parts of one", p.CounterCounts)
	}
	kinds := map[string]bool{}
	for _, k := range p.CounterKinds {
		kinds[k] = true
	}
	if !kinds["+1/+1"] || !kinds[game.CounterShield] {
		t.Errorf("counter_kinds = %v, want the +1/+1 and the shield", p.CounterKinds)
	}
	dispatchAll(t, g, active.ID, moves)
}

// One kind that covers the cost: one part, counter_kind, no array.
func TestSelfAnyKindsCounterCostSendsOneKindWhenOneKindPays(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	src := counterCostCard("Self Any Kinds Test", selfAnyKindsCost())
	src.Counters = map[string]int{game.CounterMinusOne: 3}
	elm := battlefieldCard(g, active, src)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := activationsOf(legal.EnumerateFor(g, active.ID), elm)
	if len(moves) != 1 {
		t.Fatalf("want exactly one payment, got %v", labels(moves))
	}
	p := anyKindParamsOf(t, moves[0])
	if p.CounterKind != game.CounterMinusOne || len(p.CounterKinds) != 0 {
		t.Errorf("counter_kind = %q, counter_kinds = %v, want -1/-1 and no array", p.CounterKind, p.CounterKinds)
	}
	if len(p.CounterCounts) != 1 || p.CounterCounts[0] != 2 {
		t.Errorf("counter_counts = %v, want [2]", p.CounterCounts)
	}
	dispatchAll(t, g, active.ID, moves)
}
