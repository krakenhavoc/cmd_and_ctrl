package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_cost_rest_test.go — the enumerator half of #789: counter
// costs on MANA abilities, a variable count, and a removal split
// across permanents.
//
// The invariant is #625's and #544's, one component wider: offer
// exactly the payments the engine accepts, never one it refuses, and
// nothing at all when nothing can pay. dispatchAll is the assertion
// that matters — every move it enumerates is sent through the real
// dispatcher against a clone.

type counterRestParams struct {
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterCounts    []int    `json:"counter_counts"`
	CounterKind      string   `json:"counter_kind"`
}

func counterRestParamsOf(t *testing.T, m legal.Move) counterRestParams {
	t.Helper()
	var p counterRestParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", m.Params, err)
	}
	return p
}

// vividLand is Vivid Creek: a plain tap ability and a charge-counter
// one that adds any colour.
func vividLand(charges int) game.Card {
	c := game.Card{
		Name:     "Vivid Test Land",
		TypeLine: "Land",
		ManaAbilities: []game.ManaAbilityShape{
			{TapCost: true, Produced: "{U}", Label: "Add {U}"},
			{
				TapCost:        true,
				RemoveCounters: &game.CounterRemovalCost{Counter: "charge", N: 1},
				Produced:       "{W|U|B|R|G}",
				Label:          "Add one mana of any color",
			},
		},
	}
	if charges > 0 {
		c.Counters = map[string]int{"charge": charges}
	}
	return c
}

// A mana ability's counter cost is enumerated exactly like an
// activated ability's: offered while it can be paid, gone when it
// cannot, and accepted by the dispatcher either way.
func TestManaAbilityCounterCostIsEnumeratedWhilePayable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	land := battlefieldCard(g, active, vividLand(1))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := movesFrom(legal.EnumerateFor(g, active.ID), land, legal.KindMana)
	if len(moves) != 2 {
		t.Fatalf("want both mana abilities offered, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// Spend the charge counter and the second ability stops being a
	// move — the enumerator must not offer what the engine refuses.
	spent := newTable(t)
	spentActive := spent.Seats[spent.Turn.ActiveSeat]
	clearHand(spentActive)
	empty := battlefieldCard(spent, spentActive, vividLand(0))
	advanceTo(t, spent, game.StepPrecombatMain)
	after := movesFrom(legal.EnumerateFor(spent, spentActive.ID), empty, legal.KindMana)
	if len(after) != 1 {
		t.Fatalf("with no charge counter: want only the plain ability, got %v", labels(after))
	}
	dispatchAll(t, spent, spentActive.ID, after)
}

// A variable removal is offered as ONE payment — every counter the
// permanent holds — and the wire carries the count, because there is
// no printed number for the engine to assume.
func TestVariableCounterCostEnumeratesTheLargestPayment(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	network := battlefieldCard(g, active, game.Card{
		Name:     "Mage-Ring Test Network",
		TypeLine: "Land",
		Counters: map[string]int{"storage": 3},
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:        true,
			RemoveCounters: &game.CounterRemovalCost{Counter: "storage", Variable: true},
			Produced:       "{C}",
			Label:          "Add {C} for each storage counter removed this way",
		}},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := movesFrom(legal.EnumerateFor(g, active.ID), network, legal.KindMana)
	if len(moves) != 1 {
		t.Fatalf("want one payment, got %v", labels(moves))
	}
	p := counterRestParamsOf(t, moves[0])
	if len(p.CounterCounts) != 1 || p.CounterCounts[0] != 3 {
		t.Errorf("counter_counts = %v, want [3] (every counter it holds)", p.CounterCounts)
	}
	dispatchAll(t, g, active.ID, moves)
}

// An among cost is offered as ONE payment that totals exactly N,
// drawn from the permanents holding the most first — and is not
// offered at all when the pool is short, even though every permanent
// in it holds a counter.
func TestAmongCounterCostEnumeratesOnePaymentTotallingN(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	cost := game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter: "+1/+1", N: 3, Among: true,
		From: costSpec("artifacts you control", func(c game.Card) bool { return c.IsArtifact() }),
	}}
	spider := battlefieldCard(g, active, counterCostCard("Iron Test Spider", cost))
	a := battlefieldCard(g, active, game.Card{
		Name: "Artifact A", TypeLine: "Artifact", Counters: map[string]int{"+1/+1": 1},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), spider); len(acts) != 0 {
		t.Fatalf("one counter against a printed three: want no activation, got %v", labels(acts))
	}

	b := battlefieldCard(g, active, game.Card{
		Name: "Artifact B", TypeLine: "Artifact", Counters: map[string]int{"+1/+1": 4},
	})
	moves := activationsOf(legal.EnumerateFor(g, active.ID), spider)
	if len(moves) != 1 {
		t.Fatalf("want one payment, got %v", labels(moves))
	}
	p := counterRestParamsOf(t, moves[0])
	total := 0
	for _, n := range p.CounterCounts {
		total += n
	}
	if total != 3 {
		t.Errorf("counter_counts %v total %d, want 3", p.CounterCounts, total)
	}
	if len(p.CounterSourceIDs) == 0 || p.CounterSourceIDs[0] != b.String() {
		t.Errorf("counter_source_ids = %v, want the fullest artifact %v first", p.CounterSourceIDs, b)
	}
	// The price a policy reads names every permanent it drains.
	if moves[0].Cost == nil || len(moves[0].Cost.Counters) != len(p.CounterSourceIDs) {
		t.Errorf("move cost = %+v, want one counter price per permanent", moves[0].Cost)
	}
	dispatchAll(t, g, active.ID, moves)
	_ = a
}

// A cost that ADDS a counter is enumerated while CR 118.3 lets it be
// paid, and priced as free — nothing is spent that a policy could
// weigh, which is exactly what the printed card means.
func TestAddCounterCostIsEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	druid := battlefieldCard(g, active, game.Card{
		Name: "Devoted Test Druid", TypeLine: "Creature — Elf Druid", Power: 0, Toughness: 2,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "untap",
			Cost:   game.AbilityCost{AddCounter: &game.CounterAddCost{Counter: "-1/-1", N: 1}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := activationsOf(legal.EnumerateFor(g, active.ID), druid)
	if len(moves) != 1 {
		t.Fatalf("want the untapper offered, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// uuidUnused keeps the import list honest for a file that names uuids
// only through the helpers above.
var _ = uuid.Nil
