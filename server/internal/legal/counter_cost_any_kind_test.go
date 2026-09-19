package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_cost_any_kind_test.go — the enumerator half of #943: an
// ANY-KIND removal split across permanents (Tekuthal, Inquiry
// Dominus).
//
// The invariant is #544's, unchanged: offer exactly the payments the
// engine accepts, never one it refuses, nothing at all when nothing
// can pay — and ONE payment for an among cost, because the sets
// differ only in which permanents are drained. What is new is that
// the parts may differ in KIND, so the move has to say which.
// dispatchAll is the assertion that matters: every move enumerated is
// sent through the real dispatcher against a clone.

type anyKindParams struct {
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterCounts    []int    `json:"counter_counts"`
	CounterKind      string   `json:"counter_kind"`
	CounterKinds     []string `json:"counter_kinds"`
}

func anyKindParamsOf(t *testing.T, m legal.Move) anyKindParams {
	t.Helper()
	var p anyKindParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", m.Params, err)
	}
	return p
}

// tekuthalShapedCost is "remove three counters from among other
// artifacts, creatures, and planeswalkers you control" — the kind
// left empty, which is the whole of #943.
func tekuthalShapedCost(self string) game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		N: 3, Among: true,
		From: costSpec("other artifacts, creatures, and planeswalkers you control", func(c game.Card) bool {
			return c.Name != self && (c.IsArtifact() || c.IsCreature() || c.IsPlaneswalker())
		}),
	}}
}

// The headline: with no one kind able to pay on its own, the
// enumerator still finds the payment — and names a kind per part.
func TestAnyKindAmongCounterCostEnumeratesAMixedPayment(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	tekuthal := battlefieldCard(g, active, counterCostCard("Tekuthal Test Dominus", tekuthalShapedCost("Tekuthal Test Dominus")))
	bear := battlefieldCard(g, active, game.Card{
		Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{"+1/+1": 2},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// Two counters against a printed three: not a move, even though a
	// permanent that can pay is on the board.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), tekuthal); len(acts) != 0 {
		t.Fatalf("two counters against a printed three: want no activation, got %v", labels(acts))
	}

	battlefieldCard(g, active, game.Card{
		Name: "Test Walker", TypeLine: "Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 2},
	})
	moves := activationsOf(legal.EnumerateFor(g, active.ID), tekuthal)
	if len(moves) != 1 {
		t.Fatalf("want exactly one payment, got %v", labels(moves))
	}
	p := anyKindParamsOf(t, moves[0])
	total := 0
	for _, n := range p.CounterCounts {
		total += n
	}
	if total != 3 {
		t.Errorf("counter_counts %v total %d, want 3", p.CounterCounts, total)
	}
	if len(p.CounterKinds) != len(p.CounterSourceIDs) {
		t.Fatalf("counter_kinds = %v, want one kind per permanent in %v", p.CounterKinds, p.CounterSourceIDs)
	}
	if p.CounterKind != "" {
		t.Errorf("counter_kind = %q, want it empty when the parts mix kinds", p.CounterKind)
	}
	// Fullest first, ties in battlefield order, so the Bear's two
	// +1/+1 counters lead and the walker's loyalty finishes a payment
	// neither kind could make alone.
	if len(p.CounterSourceIDs) != 2 {
		t.Fatalf("counter_source_ids = %v, want two permanents", p.CounterSourceIDs)
	}
	if p.CounterSourceIDs[0] != bear.String() || p.CounterKinds[0] != "+1/+1" {
		t.Errorf("first part = (%v, %q), want the Bear %v and its +1/+1",
			p.CounterSourceIDs[0], p.CounterKinds[0], bear)
	}
	if p.CounterKinds[1] != game.CounterLoyalty {
		t.Errorf("second part kind = %q, want loyalty", p.CounterKinds[1])
	}
	// The price a policy reads names the kind each permanent pays in,
	// not one kind for the payment.
	if moves[0].Cost == nil || len(moves[0].Cost.Counters) != len(p.CounterSourceIDs) {
		t.Fatalf("move cost = %+v, want one counter price per permanent", moves[0].Cost)
	}
	kinds := map[string]bool{}
	for _, c := range moves[0].Cost.Counters {
		kinds[c.Counter] = true
	}
	if !kinds[game.CounterLoyalty] || !kinds["+1/+1"] {
		t.Errorf("move cost counters = %+v, want both kinds priced", moves[0].Cost.Counters)
	}
	dispatchAll(t, g, active.ID, moves)
}

// When one kind covers the payment the move speaks the older wire
// shape — counter_kind, no array — so a client that predates #943
// reads the same bytes it always did.
func TestAnyKindAmongCounterCostSendsOneKindWhenOneKindPays(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	tekuthal := battlefieldCard(g, active, counterCostCard("Tekuthal Test Dominus", tekuthalShapedCost("Tekuthal Test Dominus")))
	battlefieldCard(g, active, game.Card{
		Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{"+1/+1": 4},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := activationsOf(legal.EnumerateFor(g, active.ID), tekuthal)
	if len(moves) != 1 {
		t.Fatalf("want exactly one payment, got %v", labels(moves))
	}
	p := anyKindParamsOf(t, moves[0])
	if p.CounterKind != "+1/+1" {
		t.Errorf("counter_kind = %q, want \"+1/+1\"", p.CounterKind)
	}
	if p.CounterKinds != nil {
		t.Errorf("counter_kinds = %v, want it omitted for a payment of one kind", p.CounterKinds)
	}
	dispatchAll(t, g, active.ID, moves)
}
