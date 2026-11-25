package game

import (
	"testing"

	"github.com/google/uuid"
)

// remove_supertypes_test.go — #6199: "is no longer snow" (CR 613.1d,
// layer 4; CR 205.4g: a permanent without the snow supertype is nonsnow,
// whatever its name). Melting's "All lands are no longer snow" and
// Arcum's Weathervane's two abilities.

func registerSupertypeMod(t *testing.T, g *Game, land uuid.UUID, ts int64, m Mod) {
	t.Helper()
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, land)
		if !g.registerScopedEffectLocked(uuid.Nil, []AffectedObject{PinObject(land, c.EnteredBattlefieldAt)},
			[]Mod{m}, IndefiniteDuration(), "test", ts) {
			t.Fatal("registration refused")
		}
		g.RecomputeLayersIfStaleLocked()
	})
}

func isSnow(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	snow := false
	g.ReadSnapshot(func() { snow = findBattlefieldCard(g, id).HasSupertype("snow") })
	return snow
}

// A snow land that is no longer snow can't pay {S}; it stays a Basic
// Forest.
func TestNoLongerSnowLandCantPaySnow(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := snowForestForTest(g, me)
	registerSupertypeMod(t, g, land, timeNowUnixNano(), RemoveSupertypesMod("Snow"))
	if isSnow(t, g, land) {
		t.Fatal("the land is still snow")
	}
	g.ReadSnapshot(func() {
		c := findBattlefieldCard(g, land)
		if !c.HasSupertype("basic") || !c.HasSubtype("Forest") {
			t.Errorf("the land lost more than snow: %+v", c.Effective())
		}
	})
	if _, ok := g.AutoTapForCost(me.ID, costFor(t, "{S}"), 0); ok {
		t.Fatal("planned an {S} off a land that is no longer snow")
	}
}

// CR 613.7: in layer 4, timestamp order. "Becomes snow" after "no longer
// snow" makes it snow; before it, the removal wins.
func TestRemoveAndAddSupertypeApplyInTimestampOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	later := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	earlier := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	t0 := timeNowUnixNano()
	registerSupertypeMod(t, g, later, t0, RemoveSupertypesMod("Snow"))
	registerSupertypeMod(t, g, later, t0+10, AddSupertypesMod("Snow"))
	registerSupertypeMod(t, g, earlier, t0+20, AddSupertypesMod("Snow"))
	registerSupertypeMod(t, g, earlier, t0+30, RemoveSupertypesMod("Snow"))
	if !isSnow(t, g, later) {
		t.Error("becomes snow after no longer snow: want snow")
	}
	if isSnow(t, g, earlier) {
		t.Error("no longer snow after becomes snow: want nonsnow")
	}
}

func TestRemoveSupertypesModProblem(t *testing.T) {
	if p := becomeNamedModProblem(RemoveSupertypesMod("Snow")); p != "" {
		t.Errorf("sound mod: %q", p)
	}
	if p := becomeNamedModProblem(RemoveSupertypesMod()); p == "" {
		t.Error("a mod naming no supertype passed")
	}
	if p := becomeNamedModProblem(RemoveSupertypesMod("Forest")); p == "" {
		t.Error("a mod naming a subtype passed")
	}
}
