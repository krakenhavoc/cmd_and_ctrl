package legal_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_cost_test.go — the enumerator half of #625's counter-removal
// cost. The invariant is #544's: offer exactly the payments the engine
// accepts, and nothing when there is none. dispatchAll is the
// assertion that matters; the rest checks the moves carry the choice
// and the price a policy needs.

func costSpec(label string, ok func(c game.Card) bool) *game.TargetSpec {
	return &game.TargetSpec{
		Mode: "permanent", Label: label, Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return ok(c) },
		Min:    1, Max: 1,
	}
}

func counterCostCard(name string, cost game.AbilityCost) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "remove counters: do nothing",
			Cost:   cost,
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

type counterParams struct {
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterKind      string   `json:"counter_kind"`
}

func counterParamsOf(t *testing.T, m legal.Move) counterParams {
	t.Helper()
	var p counterParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", m.Params, err)
	}
	return p
}

func walkerCard(name string, loyalty int) game.Card {
	return game.Card{
		Name: name, TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	}
}

// Heart of Kiran's shape: not offered with no walker, one move per
// walker that can pay, most loyalty first, each priced against the
// walker it charges and each accepted by the dispatcher.
func TestCounterCostIsOfferedPerPayingPlaneswalker(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	cost := game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter: game.CounterLoyalty, N: 1,
		From: costSpec("a planeswalker you control", func(c game.Card) bool { return c.IsPlaneswalker() }),
	}}
	heart := battlefieldCard(g, active, counterCostCard("Heart", cost))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), heart); len(acts) != 0 {
		t.Fatalf("no planeswalker: want no activation, got %v", labels(acts))
	}

	battlefieldCard(g, opp, walkerCard("Their Walker", 5))
	battlefieldCard(g, active, walkerCard("Empty Walker", 0))
	battlefieldCard(g, active, game.Card{Name: "Loyal Rock", TypeLine: "Artifact", Counters: map[string]int{game.CounterLoyalty: 9}})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), heart); len(acts) != 0 {
		t.Fatalf("only unpayable candidates: want no activation, got %v", labels(acts))
	}

	small := battlefieldCard(g, active, walkerCard("Small Walker", 1))
	big := battlefieldCard(g, active, walkerCard("Big Walker", 4))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, heart)
	if len(acts) != 2 {
		t.Fatalf("want one activation per paying walker, got %v", labels(acts))
	}
	for i, want := range []uuid.UUID{big, small} {
		p := counterParamsOf(t, acts[i])
		if len(p.CounterSourceIDs) != 1 || p.CounterSourceIDs[0] != want.String() {
			t.Errorf("move %d names %v, want %s (most loyalty first)", i, p.CounterSourceIDs, want)
		}
		if p.CounterKind != "" {
			t.Errorf("move %d sends counter_kind %q for a cost that prints its kind", i, p.CounterKind)
		}
		c := acts[i].Cost
		if c == nil || len(c.Counters) != 1 || c.Counters[0].CardID != want || c.Counters[0].Counter != game.CounterLoyalty || c.Counters[0].N != 1 {
			t.Errorf("move %d cost %+v, want one loyalty counter from %s", i, c, want)
		}
		if c != nil && c.Loyalty != 0 {
			t.Errorf("move %d prices a loyalty activation on the source: %+v", i, c)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// Fain's shape: "a counter" of any kind is one move per (creature,
// kind), with the kind on the wire.
func TestAnyKindCounterCostIsOfferedPerKind(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	cost := game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		N:    1,
		From: costSpec("a creature you control", func(c game.Card) bool { return c.IsCreature() }),
	}}
	fain := battlefieldCard(g, active, counterCostCard("Broker", cost))
	bear := creature("Bear", "{1}{G}", 2, 2)
	bear.Counters = map[string]int{game.CounterPlusOne: 2, game.CounterStun: 1}
	bearID := battlefieldCard(g, active, bear)
	battlefieldCard(g, active, creature("Bare Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, fain)
	if len(acts) != 2 {
		t.Fatalf("want one move per kind on the bear, got %v", labels(acts))
	}
	wantKinds := []string{game.CounterPlusOne, game.CounterStun}
	for i, m := range acts {
		p := counterParamsOf(t, m)
		if len(p.CounterSourceIDs) != 1 || p.CounterSourceIDs[0] != bearID.String() || p.CounterKind != wantKinds[i] {
			t.Errorf("move %d params %+v, want the bear and %q", i, p, wantKinds[i])
		}
		if m.Cost == nil || len(m.Cost.Counters) != 1 || m.Cost.Counters[0].Counter != wantKinds[i] {
			t.Errorf("move %d cost %+v", i, m.Cost)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// The self form names nothing on the wire and is offered only while the
// source holds the counters.
func TestSelfCounterCostIsOfferedOnlyWithCounters(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	cost := game.AbilityCost{Tap: true, RemoveCounters: &game.CounterRemovalCost{Counter: "gold", N: 1}}
	hoard := battlefieldCard(g, active, counterCostCard("Hoard", cost))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), hoard); len(acts) != 0 {
		t.Fatalf("no gold counter: want no activation, got %v", labels(acts))
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == hoard {
			g.Battlefield.Cards[i].Counters = map[string]int{"gold": 1}
		}
	}
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, hoard)
	if len(acts) != 1 {
		t.Fatalf("want one activation, got %v", labels(acts))
	}
	if p := counterParamsOf(t, acts[0]); len(p.CounterSourceIDs) != 0 || p.CounterKind != "" {
		t.Errorf("the self form sends nothing: %+v", p)
	}
	if c := acts[0].Cost; c == nil || len(c.Counters) != 1 || c.Counters[0].CardID != hoard || c.Counters[0].Counter != "gold" {
		t.Errorf("cost %+v, want one gold counter from the Hoard itself", c)
	}
	dispatchAll(t, g, active.ID, moves)
}

// The expansion budget caps the payments, and the ones it keeps are
// the cheapest: the walkers with the most loyalty.
func TestCounterCostExpansionKeepsTheMostCounters(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	cost := game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter: game.CounterLoyalty, N: 1,
		From: costSpec("a planeswalker you control", func(c game.Card) bool { return c.IsPlaneswalker() }),
	}}
	heart := battlefieldCard(g, active, counterCostCard("Heart", cost))
	for i := 1; i <= 6; i++ {
		battlefieldCard(g, active, walkerCard(fmt.Sprintf("Walker %d", i), i))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateForWithOptions(g, active.ID, legal.Options{MaxExpansionPerSource: 3})
	acts := activationsOf(moves, heart)
	if len(acts) != 3 {
		t.Fatalf("budget 3: got %d moves %v", len(acts), labels(acts))
	}
	for i, want := range []string{"Walker 6", "Walker 5", "Walker 4"} {
		if got := acts[i].Label; !strings.Contains(got, want) {
			t.Errorf("move %d = %q, want it to name %s", i, got, want)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}
