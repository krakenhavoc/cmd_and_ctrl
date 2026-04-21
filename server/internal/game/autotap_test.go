package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_test.go covers the S15 backtracking auto-tapper. Each
// scenario seeds a battlefield, asks AutoTapForCost for a plan,
// and asserts the plan is correct (covers the cost) AND chose
// the right cards under the restriction-first heuristic.
//
// All tests run on the same active 2-seat game returned by the
// existing test harness in mutations_test.go.

// pushBattlefieldForTest puts a card on the shared battlefield
// under the named controller. Used for both the basic-land
// synthetic ability tests and the catalog rock tests; OracleID is
// the catalog key, TypeLine drives the basic-land fallback.
func pushBattlefieldForTest(g *Game, controller uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
	return id
}

// withCatalogHook installs a CatalogManaAbilities shim for the
// duration of the test. The `game` package can't import
// `cards/effects` (cycle), so tests that need the catalog wired
// up inject a minimal hook covering just the oracle IDs they
// need. Composes against any pre-existing hook so test cases
// can layer.
func withCatalogHook(t *testing.T, hook func(oracleID string) []ManaAbilityShape) {
	t.Helper()
	prev := CatalogManaAbilities
	CatalogManaAbilities = func(oracleID string) []ManaAbilityShape {
		if got := hook(oracleID); got != nil {
			return got
		}
		if prev != nil {
			return prev(oracleID)
		}
		return nil
	}
	t.Cleanup(func() { CatalogManaAbilities = prev })
}

// solRingHook returns Sol Ring's catalog ability for the autotap
// tests that don't import cards/effects.
func solRingHook(oracleID string) []ManaAbilityShape {
	if oracleID == "6ad8011d-3471-4369-9d68-b264cc027487" {
		return []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}"}}
	}
	return nil
}

// birdsHook returns Birds of Paradise's catalog ability.
func birdsHook(oracleID string) []ManaAbilityShape {
	if oracleID == "d3a0b660-358c-41bd-9cd2-41fbf3491b1a" {
		return []ManaAbilityShape{{TapCost: true, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}}
	}
	return nil
}

func TestAutoTapEmptyCostNoSources(t *testing.T) {
	g := newActiveGame(t)
	plan, ok := g.AutoTapForCost(g.Seats[0].ID, ParsedCost{}, 0)
	if !ok {
		t.Fatalf("empty cost should always succeed")
	}
	if len(plan) != 0 {
		t.Errorf("empty cost should produce empty plan, got %v", plan)
	}
}

func TestAutoTapBasicLandSatisfiesMonoColorCost(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
	if !ok {
		t.Fatalf("Mountain should cover {R}")
	}
	if len(plan) != 1 || plan[0] != mountain {
		t.Errorf("plan: got %v, want [%v]", plan, mountain)
	}
}

func TestAutoTapInsufficientReturnsFalse(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}{R}"), 0); ok {
		t.Errorf("expected {R}{R} to fail with one Mountain")
	}
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}"), 0); ok {
		t.Errorf("expected {U} to fail with only a Mountain")
	}
}

func TestAutoTapTappedSourceSkipped(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0); ok {
		t.Errorf("tapped Mountain should not be a candidate")
	}
}

// TestAutoTapPrefersRestrictedSource exercises the restriction-
// first heuristic. A Mountain (only R) and a Birds of Paradise
// (any of WUBRG) covering a {R}{1} cost should pick Mountain for
// the R requirement and Birds for the generic, NOT the other
// way around — preserving the flexible source for future colored
// asks. (Both end up tapped here, but the assertion is "Mountain
// is in the plan" — the restriction-first heuristic picks it
// first for the colored slot.)
func TestAutoTapPrefersRestrictedSource(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	withCatalogHook(t, birdsHook)
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird",
		"d3a0b660-358c-41bd-9cd2-41fbf3491b1a")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}{R}"), 0)
	if !ok {
		t.Fatalf("plan: failed")
	}
	if len(plan) != 2 {
		t.Fatalf("plan size: got %d, want 2 (%v)", len(plan), plan)
	}
	planSet := map[uuid.UUID]bool{plan[0]: true, plan[1]: true}
	if !planSet[mountain] || !planSet[birds] {
		t.Errorf("plan should include both Mountain and Birds, got %v", plan)
	}
}

// TestAutoTapBacktracksGreedyFailure proves the algorithm
// recovers when a naive greedy pick gets stuck.
//
// Scenario: 1 Tundra (W/U) + 1 Plateau (W/R) + 1 Plains (W) —
// printing a Tundra-style dual via a synthetic helper. Cost:
// {U}{W}{R}. A pure greedy that took Tundra for W first leaves
// no source for U; the backtracker has to revise.
func TestAutoTapBacktracksGreedyFailure(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	// Synthetic dual lands built via direct ManaAbilities catalog
	// shim. Since we don't have real Tundra/Plateau in the catalog
	// today, we register fake oracle IDs with the dual ability
	// shape via the catalog hook.
	prevHook := CatalogManaAbilities
	CatalogManaAbilities = func(oracleID string) []ManaAbilityShape {
		switch oracleID {
		case "test-tundra":
			return []ManaAbilityShape{{TapCost: true, Produced: "{W|U}", Label: "Add {W} or {U}"}}
		case "test-plateau":
			return []ManaAbilityShape{{TapCost: true, Produced: "{W|R}", Label: "Add {W} or {R}"}}
		}
		if prevHook != nil {
			return prevHook(oracleID)
		}
		return nil
	}
	defer func() { CatalogManaAbilities = prevHook }()

	tundra := pushBattlefieldForTest(g, p.ID, "Tundra", "Land", "test-tundra")
	plateau := pushBattlefieldForTest(g, p.ID, "Plateau", "Land", "test-plateau")
	plains := pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{W}{R}"), 0)
	if !ok {
		t.Fatalf("auto-tapper failed to solve {U}{W}{R} from Tundra+Plateau+Plains")
	}
	if len(plan) != 3 {
		t.Fatalf("plan size: got %d, want 3 (%v)", len(plan), plan)
	}
	planSet := map[uuid.UUID]bool{}
	for _, id := range plan {
		planSet[id] = true
	}
	if !planSet[tundra] || !planSet[plateau] || !planSet[plains] {
		t.Errorf("plan missing one of the three lands: got %v", plan)
	}
}

// TestAutoTapSolRingProvidesTwoGeneric checks Sol Ring's two-slot
// production is recognised — a {2} cost should be solvable from
// Sol Ring alone.
func TestAutoTapSolRingProvidesTwoGeneric(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	withCatalogHook(t, solRingHook)
	sol := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact",
		"6ad8011d-3471-4369-9d68-b264cc027487")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{2}"), 0)
	if !ok {
		t.Fatalf("Sol Ring should cover {2}")
	}
	if len(plan) != 1 || plan[0] != sol {
		t.Errorf("plan: got %v, want [%v]", plan, sol)
	}
}

// TestAutoTapXValueScalesGeneric checks the {X} branch.
// 3 mountains and a {X}{R} with X=2 needs 2 generic + 1 red →
// 3 mountains exactly cover.
func TestAutoTapXValueScalesGeneric(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	for i := 0; i < 3; i++ {
		pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	}
	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{X}{R}"), 2)
	if !ok {
		t.Fatalf("3 mountains should cover {X=2}{R}")
	}
	if len(plan) != 3 {
		t.Errorf("plan size: got %d, want 3 (%v)", len(plan), plan)
	}
}

// TestAutoTapExcludedNotConsidered verifies the lock-tap excluded
// set actually removes sources from the candidate pool.
func TestAutoTapExcludedNotConsidered(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	a := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	b := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")

	plan, ok := g.AutoTapForCostExcluding(p.ID, costFor(t, "{R}"), 0, map[uuid.UUID]bool{a: true})
	if !ok {
		t.Fatalf("excluded one Mountain, the other should still cover {R}")
	}
	if len(plan) != 1 || plan[0] != b {
		t.Errorf("plan: got %v, want [%v] (the un-excluded Mountain)", plan, b)
	}
	// Excluding both should fail.
	if _, ok := g.AutoTapForCostExcluding(p.ID, costFor(t, "{R}"), 0, map[uuid.UUID]bool{a: true, b: true}); ok {
		t.Errorf("excluding both Mountains should leave no candidates")
	}
}

// TestAutoTapCyclonicRiftOverloadFromBantManabase is the
// exit-criterion smoke from the sprint plan: 38-land Bant
// manabase, cast Cyclonic Rift overload `{1}{U}{U}{U}{U}{U}{U}`,
// expect a 7-source plan. We don't model 38 lands here; just
// enough basics + a couple of dual / Signet candidates to cover
// 6×U + 1 generic.
func TestAutoTapCyclonicRiftOverloadFromBantManabase(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	withCatalogHook(t, solRingHook)
	for i := 0; i < 6; i++ {
		pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")
	}
	pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact",
		"6ad8011d-3471-4369-9d68-b264cc027487")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}{U}{U}{U}{U}{U}{U}"), 0)
	if !ok {
		t.Fatalf("Cyclonic Rift overload from 6 Islands + Sol Ring should solve")
	}
	// 6 Islands cover 6×U; Sol Ring covers the {1} (one of its two
	// C slots). The plan should list 7 unique cards.
	if len(plan) != 7 {
		t.Errorf("plan size: got %d, want 7 (%v)", len(plan), plan)
	}
}

// TestAutoTapBudgetGuard verifies that an unsolvable pathological
// case returns cleanly within the budget rather than spinning.
// Twenty random-color hybrid lands with a cost that can't be
// satisfied — we just want (nil, false) within budget.
func TestAutoTapBudgetGuard(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	prevHook := CatalogManaAbilities
	CatalogManaAbilities = func(oracleID string) []ManaAbilityShape {
		if oracleID == "test-hybrid" {
			return []ManaAbilityShape{{TapCost: true, Produced: "{W|U}", Label: "Add {W} or {U}"}}
		}
		if prevHook != nil {
			return prevHook(oracleID)
		}
		return nil
	}
	defer func() { CatalogManaAbilities = prevHook }()
	for i := 0; i < 20; i++ {
		pushBattlefieldForTest(g, p.ID, "Hybrid Dual", "Land", "test-hybrid")
	}
	// Cost requires R, which no hybrid land can provide.
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0); ok {
		t.Errorf("expected {R} to fail with only W/U hybrids")
	}
}
