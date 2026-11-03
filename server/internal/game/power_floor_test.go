package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// power_floor_test.go — #4678: "can't attack or block unless its power
// is 6 or greater" is judged on the FINISHED power (CR 613.11), so a
// later layer's anthem or counters lift it and a -1/-1 counter brings it
// back down.

const (
	testPowerFloorOracle  = "test-power-floor-6"
	testPowerAnthemOracle = "test-power-floor-anthem"
)

// withPowerFloorStatics gives testPowerFloorOracle the catalog
// constructor's layer-6 floor and testPowerAnthemOracle a layer-7c
// "creatures you control get +3/+0".
func withPowerFloorStatics(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	CatalogStaticAbilities = func(key string) []StaticAbility {
		switch key {
		case testPowerFloorOracle:
			return []StaticAbility{{
				Layer:     Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.AttackBlockPowerFloor = 6 },
			}}
		case testPowerAnthemOracle:
			return []StaticAbility{{
				Layer:    Layer7PT,
				SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Power += 3 },
			}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
}

func pushPowerFloorCard(t *testing.T, g *Game, owner *Player, name, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := pushCombatant(t, g, owner, name, power, toughness)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].OracleID = oracle
			}
		}
		g.layerVersion.Add(1)
	})
	return id
}

func powerFloorRestricted(t *testing.T, g *Game, id uuid.UUID) (cantAttack, cantBlock bool) {
	t.Helper()
	ch := scopedEffectChar(t, g, id)
	return ch.Restrictions.Has(CantAttack), ch.Restrictions.Has(CantBlock)
}

func TestPowerFloorReadsCountersAfterTheLayers(t *testing.T) {
	withPowerFloorStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	drome := pushPowerFloorCard(t, g, me, "Technodrome", testPowerFloorOracle, 3, 3)
	plain := pushCombatant(t, g, me, "Bear", 2, 2)

	if a, b := powerFloorRestricted(t, g, drome); !a || !b {
		t.Fatalf("a 3/3 can attack %v / block %v, want neither", !a, !b)
	}
	if a, b := powerFloorRestricted(t, g, plain); a || b {
		t.Fatal("the floor reached a creature that doesn't have it")
	}

	addCounterForTest(t, g, drome, CounterPlusOne, 3)
	if a, b := powerFloorRestricted(t, g, drome); a || b {
		t.Fatalf("at power 6 it still can't attack %v / block %v", a, b)
	}

	// -1/-1 annihilates one +1/+1 (CR 704.5q); either way it is a 5/5.
	addCounterForTest(t, g, drome, CounterMinusOne, 1)
	if a, b := powerFloorRestricted(t, g, drome); !a || !b {
		t.Fatalf("back at power 5 it can attack %v / block %v", !a, !b)
	}
}

func TestPowerFloorCountsAnAnthemFromALaterLayer(t *testing.T) {
	withPowerFloorStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	drome := pushPowerFloorCard(t, g, me, "Technodrome", testPowerFloorOracle, 3, 3)
	pushPowerFloorCard(t, g, me, "Anthem", testPowerAnthemOracle, 0, 1)
	if a, b := powerFloorRestricted(t, g, drome); a || b {
		t.Fatalf("a layer-7c +3/+0 to power 6 left can't attack %v / can't block %v", a, b)
	}
}

func TestPowerFloorRefusesTheDeclarationsAndLiftsThem(t *testing.T) {
	withPowerFloorStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drome := pushPowerFloorCard(t, g, me, "Technodrome", testPowerFloorOracle, 3, 3)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(drome, opp.ID); !errors.Is(err, ErrCantAttack) {
		t.Fatalf("a 3/3 attacked: err = %v, want ErrCantAttack", err)
	}
	addCounterForTest(t, g, drome, CounterPlusOne, 3)
	if err := g.DeclareAttacker(drome, opp.ID); err != nil {
		t.Fatalf("a 6/6 can't attack: %v", err)
	}
	// CR 508.1c: checked at declaration only. Shrinking it now does not
	// take it out of combat.
	addCounterForTest(t, g, drome, CounterMinusOne, 1)
	var attacking bool
	g.ReadSnapshot(func() {
		if c, ok := g.battlefieldCardLocked(drome); ok {
			attacking = c.AttackingTarget != uuid.Nil
		}
	})
	if !attacking {
		t.Error("shrinking a declared attacker below the floor removed it from combat")
	}
}

func TestPowerFloorRefusesABlock(t *testing.T) {
	withPowerFloorStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Bear", 2, 2)
	drome := pushPowerFloorCard(t, g, opp, "Technodrome", testPowerFloorOracle, 3, 3)
	var a, b *Card
	ok := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, b = findBattlefieldCard(g, attacker), findBattlefieldCard(g, drome)
		ok = g.CanBlockLocked(a, b)
	})
	if ok {
		t.Fatal("a 3/3 may block")
	}
	addCounterForTest(t, g, drome, CounterPlusOne, 3)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, b = findBattlefieldCard(g, attacker), findBattlefieldCard(g, drome)
		ok = g.CanBlockLocked(a, b)
	})
	if !ok {
		t.Fatal("a 6/6 may not block")
	}
}
