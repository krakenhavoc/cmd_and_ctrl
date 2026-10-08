package effects

import (
	"testing"
)

// gilded_drake_test.go pins #2182: "this ability still resolves if
// its target becomes illegal" (Gilded Drake, CR 608.2b).

const gildedDrakeOracle = "7f06c098-6482-4bf3-a9a1-110d6d5b5703"

func TestGildedDrakeExchangesWithTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	drake := castAndResolveCreature(t, g, "Gilded Drake", "Creature - Drake", gildedDrakeOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("target controller %s, want %s", got, me.ID)
	}
	if got := controllerOf(t, g, drake); got != opp.ID {
		t.Errorf("Drake controller %s, want %s", got, opp.ID)
	}
}

// The seam: the only target dies in response, the trigger still
// resolves, no exchange happens and the Drake is sacrificed.
func TestGildedDrakeIsSacrificedWhenTheTargetBecomesIllegal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	drake := castAndResolveCreature(t, g, "Gilded Drake", "Creature - Drake", gildedDrakeOracle)
	pickCard(t, g, me.ID, victim)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)

	if battlefieldHas(g, drake) {
		t.Errorf("the Drake stayed on the battlefield after its target left: the trigger was removed under CR 608.2b")
	}
	if !inGraveyardOf(g, me.ID, drake) {
		t.Errorf("the sacrificed Drake is not in its owner's graveyard")
	}
}

// "Up to one": with nothing to point at, the Drake still triggers and
// is sacrificed.
func TestGildedDrakeIsSacrificedWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	drake := castAndResolveCreature(t, g, "Gilded Drake", "Creature - Drake", gildedDrakeOracle)
	passPriorityAroundTable(t, g)

	if battlefieldHas(g, drake) {
		t.Errorf("a Drake with no target stayed on the battlefield")
	}
	if !inGraveyardOf(g, me.ID, drake) {
		t.Errorf("the sacrificed Drake is not in its owner's graveyard")
	}
}
