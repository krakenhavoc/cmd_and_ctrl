package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_generic_order_test.go — #2278. A generic pip is paid by the
// source that makes the FEWEST colours, so what stays untapped can make
// the most. A Plains is spent before a Glacial Fortress.

const (
	testDualOracle  = "test-2278-dual-wu"
	testTriOracle   = "test-2278-tri-wub"
	testRiderOracle = "test-2392-rider"
	testOpaqueRider = "test-2392-opaque-rider"
)

func genericOrderHook(oracleID string) []ManaAbilityShape {
	switch oracleID {
	case testDualOracle:
		return []ManaAbilityShape{{TapCost: true, Produced: "{W|U}", Label: "Add {W} or {U}"}}
	case testTriOracle:
		return []ManaAbilityShape{{TapCost: true, Produced: "{W|U|B}", Label: "Add {W}, {U} or {B}"}}
	}
	return nil
}

func planSet(plan []uuid.UUID) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(plan))
	for _, id := range plan {
		out[id] = true
	}
	return out
}

// The test the issue asks for: {2} with a Plains, a Glacial Fortress and
// an Island untapped spends the Plains and the Island and leaves the
// Fortress.
func TestAutoTapGenericSpendsBasicsBeforeADual(t *testing.T) {
	withCatalogHook(t, genericOrderHook)
	g := newActiveGame(t)
	p := g.Seats[0]
	plains := pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	fortress := pushBattlefieldForTest(g, p.ID, "Glacial Fortress", "Land", testDualOracle)
	island := pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{2}"), 0)
	if !ok || len(plan) != 2 {
		t.Fatalf("plan = %v ok=%v, want two sources", plan, ok)
	}
	got := planSet(plan)
	if !got[plains] || !got[island] || got[fortress] {
		t.Errorf("plan = %v, want the Plains (%v) and the Island (%v), not the Fortress (%v)", plan, plains, island, fortress)
	}
}

// Among multi-colour sources, the one with fewer colours goes first:
// a dual before a tri-land.
func TestAutoTapGenericSpendsADualBeforeATriLand(t *testing.T) {
	withCatalogHook(t, genericOrderHook)
	g := newActiveGame(t)
	p := g.Seats[0]
	dual := pushBattlefieldForTest(g, p.ID, "Glacial Fortress", "Land", testDualOracle)
	tri := pushBattlefieldForTest(g, p.ID, "Raffine's Tower", "Land", testTriOracle)

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != dual {
		t.Errorf("plan = %v ok=%v, want the dual (%v), not the tri-land (%v)", plan, ok, dual, tri)
	}
}

// Colourless is still first (unchanged): a Sol Ring before a basic.
func TestAutoTapGenericStillSpendsColorlessFirst(t *testing.T) {
	withCatalogHook(t, solRingHook)
	g := newActiveGame(t)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{2}"), 0)
	if !ok || len(plan) != 1 || plan[0] != ring {
		t.Errorf("plan = %v ok=%v, want Sol Ring alone", plan, ok)
	}
}

// The tier key, pinned.
func TestTierForGenericCountsColors(t *testing.T) {
	for _, tc := range []struct {
		slots []ProducedManaEntry
		want  int
	}{
		{[]ProducedManaEntry{{Options: []string{"C"}}, {Options: []string{"C"}}}, 0},
		{[]ProducedManaEntry{{Options: []string{"W"}}}, 1},
		{[]ProducedManaEntry{{Options: []string{"W", "U"}}}, 2},
		{[]ProducedManaEntry{{Options: []string{"W"}}, {Options: []string{"U"}}}, 2},
		{[]ProducedManaEntry{{Options: []string{"W", "U", "B", "R", "G"}}}, 5},
	} {
		if got := tierForGeneric(tapSource{Slots: tc.slots}); got != tc.want {
			t.Errorf("tierForGeneric(%v) = %d, want %d", tc.slots, got, tc.want)
		}
	}
}

// #2392: a rider is planned only when it declares what it deals. An
// opaque one is a closure the planner cannot price.
func TestAutoTapPlansOnlyADeclaredRider(t *testing.T) {
	rider := func(*Game, uuid.UUID, uuid.UUID) error { return nil }
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		switch oracleID {
		case testRiderOracle:
			return []ManaAbilityShape{{TapCost: true, Produced: "{R}", Label: "Add {R}. 1 damage", Rider: rider, RiderSelfDamage: 1}}
		case testOpaqueRider:
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}. Something", Rider: rider}}
		}
		return nil
	})
	g := newActiveGame(t)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Pain Source", "Land", testRiderOracle)
	pushBattlefieldForTest(g, p.ID, "Opaque Source", "Land", testOpaqueRider)

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0); !ok {
		t.Error("a declared 1-damage rider is not planned")
	}
	if plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{G}"), 0); ok {
		t.Errorf("an opaque rider was planned: %v", plan)
	}
}
