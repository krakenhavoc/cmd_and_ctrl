package game

import (
	"errors"
	"testing"
)

// autotap_energy_test.go — ADR 0129 §5, owner decision 2: a mana
// ability that pays energy (Aether Hub) is planned in the energy tier,
// only when no plan without energy pays, before the pain tier, and the
// whole plan is held to the controller's energy.

const (
	testHubOracle  = "test-0129-hub"
	testPainOracle = "test-0129-pain"
)

func energyHook(oracleID string) []ManaAbilityShape {
	switch oracleID {
	case testHubOracle:
		return []ManaAbilityShape{
			{TapCost: true, Produced: "{C}", Label: "{T}: Add {C}."},
			{TapCost: true, EnergyCost: 1, Produced: "{W|U|B|R|G}", Label: "{T}, Pay {E}: Add one mana of any color."},
		}
	case testPainOracle:
		return []ManaAbilityShape{{TapCost: true, LifeCost: 1, Produced: "{W|U|B|R|G}", Label: "{T}, Pay 1 life: Add one mana of any color."}}
	}
	return nil
}

func energyTable(t *testing.T, energy int) (*Game, *Player) {
	t.Helper()
	withCatalogHook(t, energyHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	setEnergyForTest(t, g, p, energy)
	return g, p
}

// The coloured half is planned only when nothing that spends no energy
// pays, and never without the energy.
func TestAutoTapSpendsEnergyOnlyWhenItMust(t *testing.T) {
	g, p := energyTable(t, 1)
	plains := pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	hub := pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{W}"), 0)
	if !ok || len(plan) != 1 || plan[0] != plains {
		t.Errorf("{W}: plan = %v ok=%v, want the Plains alone", plan, ok)
	}
	plan, ok = g.AutoTapForCost(p.ID, costFor(t, "{U}"), 0)
	if !ok || len(plan) != 1 || plan[0] != hub {
		t.Errorf("{U}: plan = %v ok=%v, want the Hub", plan, ok)
	}

	setEnergyForTest(t, g, p, 0)
	if plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}"), 0); ok {
		t.Errorf("{U} with no energy was planned: %v", plan)
	}
}

// Two Hubs with one energy between them pay one coloured pip, not two.
func TestAutoTapHoldsThePlanToTheControllersEnergy(t *testing.T) {
	g, p := energyTable(t, 1)
	pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)
	pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}"), 0); !ok {
		t.Error("{U}: one energy should pay one pip")
	}
	if plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{B}"), 0); ok {
		t.Errorf("{U}{B} with one energy was planned: %v", plan)
	}
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{1}"), 0); !ok {
		t.Error("{U}{1}: the second Hub's {C} pays the generic")
	}
	setEnergyForTest(t, g, p, 2)
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{B}"), 0); !ok {
		t.Error("{U}{B} with two energy is not planned")
	}
}

// Energy before life: with both a Hub and a pain source, the Hub pays.
func TestAutoTapSpendsEnergyBeforeLife(t *testing.T) {
	g, p := energyTable(t, 1)
	pain := pushBattlefieldForTest(g, p.ID, "Mana Confluence", "Land", testPainOracle)
	hub := pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}"), 0)
	if !ok || len(plan) != 1 || plan[0] != hub {
		t.Errorf("plan = %v ok=%v, want the Hub (%v) before the pain source (%v)", plan, ok, hub, pain)
	}
}

// The cast pays the energy, and the preview lists it against the Hub.
func TestAutoTappedCastPaysTheEnergyAndThePreviewSaysSo(t *testing.T) {
	g, p := energyTable(t, 2)
	hub := pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)

	entries, ok := g.AutoTapPlanPreferringExcluding(p.ID, costFor(t, "{U}"), 0, nil, 0)
	if !ok || len(entries) != 1 || entries[0].CardID != hub || entries[0].Energy != 1 || !entries[0].Taps {
		t.Fatalf("preview = %+v ok=%v, want the Hub, tapped, paying 1 energy", entries, ok)
	}
	id := pushTypedCardToHandWithCost(p, "Blue Spell", "Instant", "{U}")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if PlayerEnergy(p) != 1 {
		t.Errorf("energy = %d after the cast, want 1", PlayerEnergy(p))
	}
}

// Activated by hand: short of energy refuses with nothing tapped; with
// it, the energy comes off and the mana lands.
func TestActivateManaAbilityPaysEnergy(t *testing.T) {
	g, p := energyTable(t, 0)
	hub := pushBattlefieldForTest(g, p.ID, "Aether Hub", "Land", testHubOracle)

	err := g.ActivateManaAbility(p.ID, hub, 1, ManaAbilityParams{Colors: []string{"U"}})
	if !errors.Is(err, ErrInsufficientEnergy) {
		t.Fatalf("no energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if c := findBattlefieldCard(g, hub); c == nil || c.Tapped {
		t.Fatal("a refused activation tapped the Hub")
	}
	setEnergyForTest(t, g, p, 1)
	if err := g.ActivateManaAbility(p.ID, hub, 1, ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if PlayerEnergy(p) != 0 {
		t.Errorf("energy = %d, want 0", PlayerEnergy(p))
	}
	if len(p.ManaPool) != 1 {
		t.Errorf("pool = %v, want one mana", p.ManaPool)
	}
}
