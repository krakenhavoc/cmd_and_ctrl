package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// power_gated_goad_test.go — #6098, Baeloth Barrityl, Entertainer:
// "creatures your opponents control with power less than ~'s power are
// goaded". Goad is an effect on the game rules, applied after every other
// continuous effect (CR 613.11), so both powers are the FINISHED ones
// (layers and P/T counters), and goad's requirements are CR 701.15b's.

const (
	testPowerGatedGoadOracle   = "test-power-gated-goad"
	testPowerGatedAnthemOracle = "test-power-gated-anthem"
)

// withPowerGatedGoadStatics gives testPowerGatedGoadOracle the catalog
// card's layer-6 record over every creature an opponent controls, and
// testPowerGatedAnthemOracle a layer-7c "creatures you control get +3/+0".
func withPowerGatedGoadStatics(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	CatalogStaticAbilities = func(key string) []StaticAbility {
		switch key {
		case testPowerGatedGoadOracle:
			return []StaticAbility{{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return target.IsCreature() && target.Controller != source.Controller
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
					c.PowerGatedGoads = append(c.PowerGatedGoads, PowerGatedGoad{Source: source.InstanceID, By: source.Controller})
				},
			}}
		case testPowerGatedAnthemOracle:
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

func powerGatedGoaders(t *testing.T, g *Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	return staticGoadersIn(scopedEffectChar(t, g, id).AttackRequirements)
}

// The bar and the creature are both read after the layers and counters:
// a 1/1 under a 2-power source is goaded by the source's controller, a 2/2
// is not, and a +1/+1 counter on the source brings the 2/2 under the bar.
func TestPowerGatedGoadComparesFinishedPowers(t *testing.T) {
	withPowerGatedGoadStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baeloth := pushPowerFloorCard(t, g, me, "Baeloth", testPowerGatedGoadOracle, 2, 5)
	small := pushCombatant(t, g, opp, "Small", 1, 1)
	even := pushCombatant(t, g, opp, "Even", 2, 2)
	mine := pushCombatant(t, g, me, "Mine", 1, 1)

	if got := powerGatedGoaders(t, g, small); !slices.Equal(got, []uuid.UUID{me.ID}) {
		t.Errorf("the 1/1 is goaded by %v, want only the source's controller", got)
	}
	if got := powerGatedGoaders(t, g, even); len(got) != 0 {
		t.Errorf("the 2/2 (power not less) is goaded by %v", got)
	}
	if got := powerGatedGoaders(t, g, mine); len(got) != 0 {
		t.Errorf("the source's controller's own creature is goaded by %v", got)
	}
	// CR 701.15b: attacks each combat if able, and a player other than
	// the goader if able — the same pair every static goad writes.
	if reqs := scopedEffectChar(t, g, small).AttackRequirements; len(reqs) != 2 || reqs[1].OtherThan != me.ID {
		t.Errorf("requirements on the 1/1 = %+v, want goad's pair", reqs)
	}

	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(baeloth, CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := powerGatedGoaders(t, g, even); !slices.Equal(got, []uuid.UUID{me.ID}) {
		t.Errorf("after a +1/+1 counter on the source the 2/2 is goaded by %v, want the source's controller", got)
	}
}

// A later layer on the creature's side lifts it over the bar: an anthem
// its controller has, then the source leaving, end the goad.
func TestPowerGatedGoadEndsWhenTheCreatureGrowsOrTheSourceLeaves(t *testing.T) {
	withPowerGatedGoadStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baeloth := pushPowerFloorCard(t, g, me, "Baeloth", testPowerGatedGoadOracle, 2, 5)
	small := pushCombatant(t, g, opp, "Small", 1, 1)
	if len(powerGatedGoaders(t, g, small)) != 1 {
		t.Fatal("setup: the 1/1 is not goaded")
	}
	anthem := pushPowerFloorCard(t, g, opp, "Anthem", testPowerGatedAnthemOracle, 0, 1)
	if got := powerGatedGoaders(t, g, small); len(got) != 0 {
		t.Errorf("the 1/1 at power 4 is still goaded by %v", got)
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, opp.Graveyard, anthem); err != nil {
			t.Fatal(err)
		}
		g.layerVersion.Add(1)
	})
	if len(powerGatedGoaders(t, g, small)) != 1 {
		t.Fatal("the anthem left, but the 1/1 is not goaded again")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, baeloth); err != nil {
			t.Fatal(err)
		}
		g.layerVersion.Add(1)
	})
	if got := powerGatedGoaders(t, g, small); len(got) != 0 {
		t.Errorf("the source has left, but the 1/1 is goaded by %v", got)
	}
}

// The goad is a goad like any other: Goaded() answers yes, and the record
// of a goaded creature that leaves remembers it (PermanentInfo.Goaded).
func TestPowerGatedGoadIsGoadedAndRemembered(t *testing.T) {
	withPowerGatedGoadStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPowerFloorCard(t, g, me, "Baeloth", testPowerGatedGoadOracle, 2, 5)
	small := pushCombatant(t, g, opp, "Small", 1, 1)
	big := pushCombatant(t, g, opp, "Big", 3, 3)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.battlefieldCardLocked(small)
		if !c.Goaded() {
			t.Error("Goaded() = false for the 1/1")
		}
		if err := g.DestroyPermanentForEffect(small); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(big); err != nil {
			t.Fatal(err)
		}
	})
	if rec, ok := g.LastKnownPermanentForEffect(small); !ok || !rec.Goaded {
		t.Errorf("the 1/1's record = %+v, %v; want Goaded", rec.Goaded, ok)
	}
	if rec, ok := g.LastKnownPermanentForEffect(big); !ok || rec.Goaded {
		t.Errorf("the 3/3's record = %+v, %v; want not Goaded", rec.Goaded, ok)
	}
}
