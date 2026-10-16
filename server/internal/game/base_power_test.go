package game

import (
	"testing"

	"github.com/google/uuid"
)

// base_power_test.go — Card.BasePower (CR 208.4b), the layer pass's
// power after layer 7b. Kutzil, Malamet Exemplar is the reader.

// basePowerOf recomputes the layers and reads the permanent's base
// power and its compared power (counters included).
func basePowerOf(t *testing.T, g *Game, id uuid.UUID) (base, power int) {
	t.Helper()
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.battlefieldCardLocked(id); ok {
			base, power, found = c.BasePower(), c.PowerForComparison(), true
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return base, power
}

func TestBasePowerIgnoresPumpsAndCountersButNotSets(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	if base, power := basePowerOf(t, g, bear); base != 2 || power != 2 {
		t.Fatalf("printed 2/2: base %d, power %d, want 2 and 2", base, power)
	}

	// A layer-7c pump raises power and leaves base power alone.
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	if base, power := basePowerOf(t, g, bear); base != 2 || power != 3 {
		t.Errorf("after +1/+1: base %d, power %d, want 2 and 3", base, power)
	}

	// A +1/+1 counter (7d) is on top of base power too.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, CounterPlusOne, 1) })
	if base, power := basePowerOf(t, g, bear); base != 2 || power != 4 {
		t.Errorf("after a counter: base %d, power %d, want 2 and 4", base, power)
	}

	// An effect that SETS base power (7b) moves base power with it.
	registerScopedEffectForTest(t, g, bear, []Mod{SetBasePowerMod(5)}, IndefiniteDuration())
	if base, power := basePowerOf(t, g, bear); base != 5 || power != 7 {
		t.Errorf("after base power 5: base %d, power %d, want 5 and 7", base, power)
	}
}

func TestBasePowerOffTheBattlefieldIsThePrintedPower(t *testing.T) {
	c := Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3}
	if got := c.BasePower(); got != 3 {
		t.Errorf("BasePower of a card with no layer pass = %d, want its printed 3", got)
	}
}
