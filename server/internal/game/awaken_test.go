package game

import (
	"testing"

	"github.com/google/uuid"
)

// awaken_test.go — ADR 0135 §3 (#2411): the engine half of awaken. The
// card half (the offer, the extra clause, Hardened Scales and Doubling
// Season against the counters-first order) is in
// cards/effects/awaken_cards_test.go.

func awakenLand(t *testing.T, g *Game, actor *Player, land uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AwakenForEffect(actor.ID, uuid.Nil, land, n); err != nil {
			t.Fatalf("AwakenForEffect(%d): %v", n, err)
		}
	})
}

// CR 702.113a: N +1/+1 counters, and a 0/0 Elemental creature with haste
// that is still a land, whose land subtype stays.
func TestAwakenMakesAHastyElementalLandCreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	awakenLand(t, g, me, land, 3)

	c := earthbentView(t, g, land)
	if !c.IsLand() || !c.IsCreature() {
		t.Errorf("land %v, creature %v; want both (CR 702.113a: \"It's still a land\")", c.IsLand(), c.IsCreature())
	}
	if !c.HasSubtype("Elemental") || !c.HasSubtype("Forest") {
		t.Errorf("subtypes: Elemental %v, Forest %v; want both", c.HasSubtype("Elemental"), c.HasSubtype("Forest"))
	}
	if !HasKeyword(c, "haste") {
		t.Error("the awakened land has no haste")
	}
	if c.Counters[CounterPlusOne] != 3 || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d counters, %d/%d; want 3 and 3/3", c.Counters[CounterPlusOne], c.CurrentPower(), c.CurrentToughness())
	}
	for _, dt := range g.DelayedTriggers {
		if dt != nil && len(dt.Cards) > 0 && dt.Cards[0] == land {
			t.Errorf("awaken scheduled a delayed trigger %q; it has no return", dt.Label)
		}
	}
}

// A land that is not on the battlefield is a silent no-op (CR 608.2b has
// already judged the target).
func TestAwakenOnALandThatLeftDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	awakenLand(t, g, me, uuid.New(), 3)
	awakenLand(t, g, me, uuid.Nil, 3)
}

// Earthbend shares the animation builder and adds no subtype.
func TestEarthbendStillAddsNoSubtype(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	earthbend(t, g, me, land, 2)
	if c := earthbentView(t, g, land); c.HasSubtype("Elemental") {
		t.Error("an earthbent land became an Elemental")
	}
}
