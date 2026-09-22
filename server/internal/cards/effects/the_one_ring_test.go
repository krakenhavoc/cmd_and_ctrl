package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theOneRingOracle = "3aa83ed2-f48b-4ce6-a614-2c54ddf50538"

// TestTheOneRingDrawsForTheNewBurdenTotal is the ordering the card
// turns on: the counter goes on FIRST, so the first activation draws
// one card, the second two, the third three.
func TestTheOneRingDrawsForTheNewBurdenTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ring := pushPermanentForTest(g, me.ID, "The One Ring", theOneRingOracle, "Legendary Artifact")

	for want := 1; want <= 3; want++ {
		before := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", want, err)
		}
		if me.Hand.Size() != before {
			t.Fatalf("activation %d drew before the ability resolved", want)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - before; got != want {
			t.Errorf("activation %d drew %d cards, want %d", want, got, want)
		}
		c, ok := g.LookupCardForEffect(ring)
		if !ok || c.Counters["burden"] != want {
			t.Errorf("after activation %d burden counters = %d, want %d", want, c.Counters["burden"], want)
		}
		// Untap it for the next activation; the card has no untap
		// clause of its own and the test is about the draw math.
		if err := g.UntapTargetForEffect(ring); err != nil {
			t.Fatalf("untap: %v", err)
		}
	}
}

// TestTheOneRingUpkeepLosesOneLifePerBurdenCounter: the drain is
// counted when the trigger resolves, not when it fires.
func TestTheOneRingUpkeepLosesOneLifePerBurdenCounter(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	ring := pushPermanentForTest(g, owner.ID, "The One Ring", theOneRingOracle, "Legendary Artifact")
	if err := g.AddCounter(ring, "burden", 3); err != nil {
		t.Fatalf("seed burden counters: %v", err)
	}
	lifeBefore := owner.Life

	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, ring) == nil {
		t.Fatalf("The One Ring's upkeep trigger is not on the stack")
	}
	if owner.Life != lifeBefore {
		t.Fatalf("the drain happened before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if owner.Life != lifeBefore-3 {
		t.Errorf("life %d -> %d, want -3 for three burden counters", lifeBefore, owner.Life)
	}
}

// TestTheOneRingWithNoBurdenCountersCostsNoLife: the turn it lands
// it has no counters, so the first upkeep is free.
func TestTheOneRingWithNoBurdenCountersCostsNoLife(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "The One Ring", theOneRingOracle, "Legendary Artifact")
	lifeBefore := owner.Life

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if owner.Life != lifeBefore {
		t.Errorf("life %d -> %d, want no change with no burden counters", lifeBefore, owner.Life)
	}
}

// TestTheOneRingIsIndestructibleAndComplete pins the keyword line and
// the completeness claim. The card carried a caveat naming the
// missing "protection from everything until your next turn" clause
// from S40 until #1197 gave a PLAYER an ability slice; now every
// clause is implemented, so the caveat is gone and the card is
// CompletenessFull. The protection itself is proved in
// the_one_ring_shield_test.go.
func TestTheOneRingIsIndestructibleAndComplete(t *testing.T) {
	spec, ok := Lookup(theOneRingOracle)
	if !ok {
		t.Fatalf("The One Ring is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("every clause of The One Ring is implemented since #1197: %+v", spec)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	ring := pushPermanentForTest(g, me.ID, "The One Ring", theOneRingOracle, "Legendary Artifact")
	c, ok := g.LookupCardForEffect(ring)
	if !ok {
		t.Fatalf("The One Ring is not on the battlefield")
	}
	if !game.HasKeyword(&c, "indestructible") {
		t.Errorf("The One Ring should be indestructible: %+v", c.Effective().Abilities)
	}
}
