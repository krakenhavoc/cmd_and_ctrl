package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// energy_cost_test.go — ADR 0129 §2: AbilityCost.Energy / EnergyX,
// paid through payEnergyLocked (CR 107.14), checked before anything is
// paid (CR 118.3, CR 601.2h). The enumerator's half is
// legal/energy_test.go and the catalog's is
// cards/effects/energy_cards_test.go.

func pushEnergySource(g *Game, p *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Energy Test", p.ID)
	c.TypeLine = "Artifact"
	c.Controller = p.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "Energy test",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, item.XValue)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func setEnergyForTest(t *testing.T, g *Game, p *Player, n int) {
	t.Helper()
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.AddPlayerCounterForEffect(p.ID, CounterEnergy, n-PlayerEnergy(p)); err != nil {
		t.Fatal(err)
	}
}

// Short of energy: refused with the sentence the view stamps, and
// nothing is paid — the {T} is not tapped and no energy comes off.
func TestEnergyCostShortRefusesAndPaysNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushEnergySource(g, me, AbilityCost{Tap: true, Energy: 3})
	setEnergyForTest(t, g, me, 2)

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrInsufficientEnergy) {
		t.Fatalf("err = %v, want ErrInsufficientEnergy", err)
	}
	if got, want := err.Error(), "Not enough energy (have 2, need 3)"; got != want {
		t.Errorf("err text = %q, want %q", got, want)
	}
	if c := findBattlefieldCard(g, src); c == nil || c.Tapped {
		t.Error("the source was tapped by a refused activation")
	}
	if PlayerEnergy(me) != 2 || me.Energy != 2 {
		t.Errorf("energy = %d (legacy %d), want 2", PlayerEnergy(me), me.Energy)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d stack items after a refusal", len(g.StackMeta))
	}
}

// Paid: the counters come off the player, the legacy int follows, and
// one EventPlayerCounterPlaced names the payer and the source with the
// negative delta.
func TestEnergyCostPaysAndEmits(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushEnergySource(g, me, AbilityCost{Energy: 3})
	setEnergyForTest(t, g, me, 5)
	before := len(g.Events)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if PlayerEnergy(me) != 2 || me.Energy != 2 {
		t.Errorf("energy = %d (legacy %d), want 2", PlayerEnergy(me), me.Energy)
	}
	n := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind != EventPlayerCounterPlaced {
			continue
		}
		n++
		if ev.Label != CounterEnergy || ev.Amount != -3 || ev.Target != me.ID || ev.Actor != me.ID || ev.Source != src {
			t.Errorf("event = %+v, want energy -3 by and on the payer from the source", ev)
		}
	}
	if n != 1 {
		t.Errorf("%d player-counter events, want 1", n)
	}
}

// Paying the whole total empties the map entry, as every player
// counter at zero does.
func TestEnergyCostPaysDownToZero(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushEnergySource(g, me, AbilityCost{Energy: 2})
	setEnergyForTest(t, g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := me.Counters[CounterEnergy]; ok || me.Energy != 0 {
		t.Errorf("counters = %v, legacy %d, want no energy entry", me.Counters, me.Energy)
	}
}

// "Pay X {E}": X is announced (CR 107.3a), may not exceed the energy
// (CR 118.3), and is what the effect reads.
func TestEnergyXIsAnnouncedAndCapped(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushEnergySource(g, me, AbilityCost{EnergyX: true})
	setEnergyForTest(t, g, me, 3)

	if !(AbilityCost{EnergyX: true}).DemandsX() {
		t.Fatal("an energy X does not demand X")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 4}); !errors.Is(err, ErrInsufficientEnergy) {
		t.Fatalf("X=4 with 3 energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("X=2: %v", err)
	}
	if PlayerEnergy(me) != 1 {
		t.Errorf("energy = %d, want 1", PlayerEnergy(me))
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want 1", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		if item.XValue != 2 {
			t.Errorf("stack item X = %d, want 2", item.XValue)
		}
	}
}

func TestAbilityEnergyCost(t *testing.T) {
	for _, tc := range []struct {
		cost AbilityCost
		x    int
		want int
	}{
		{AbilityCost{}, 3, 0},
		{AbilityCost{Energy: 8}, 0, 8},
		{AbilityCost{Energy: 8}, 3, 8},
		{AbilityCost{EnergyX: true}, 3, 3},
		{AbilityCost{Energy: 1, EnergyX: true}, 3, 4},
	} {
		if got := AbilityEnergyCost(tc.cost, tc.x); got != tc.want {
			t.Errorf("AbilityEnergyCost(%+v, %d) = %d, want %d", tc.cost, tc.x, got, tc.want)
		}
	}
}
