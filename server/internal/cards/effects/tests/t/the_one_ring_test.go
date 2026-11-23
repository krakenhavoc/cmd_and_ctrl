package testst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTheOneRingDrawsForTheNewBurdenTotal is the ordering the card
// turns on: the counter goes on FIRST, so the first activation draws
// one card, the second two, the third three.
func TestTheOneRingDrawsForTheNewBurdenTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushPermanentForTest(g, me.ID, "The One Ring", effectstest.TheOneRingOracle, "Legendary Artifact")

	for want := 1; want <= 3; want++ {
		before := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", want, err)
		}
		if me.Hand.Size() != before {
			t.Fatalf("activation %d drew before the ability resolved", want)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	ring := effectstest.PushPermanentForTest(g, owner.ID, "The One Ring", effectstest.TheOneRingOracle, "Legendary Artifact")
	if err := g.AddCounter(ring, "burden", 3); err != nil {
		t.Fatalf("seed burden counters: %v", err)
	}
	lifeBefore := owner.Life

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, ring) == nil {
		t.Fatalf("The One Ring's upkeep trigger is not on the stack")
	}
	if owner.Life != lifeBefore {
		t.Fatalf("the drain happened before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != lifeBefore-3 {
		t.Errorf("life %d -> %d, want -3 for three burden counters", lifeBefore, owner.Life)
	}
}

// TestTheOneRingWithNoBurdenCountersCostsNoLife: the turn it lands
// it has no counters, so the first upkeep is free.
func TestTheOneRingWithNoBurdenCountersCostsNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "The One Ring", effectstest.TheOneRingOracle, "Legendary Artifact")
	lifeBefore := owner.Life

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	spec, ok := effects.Lookup(effectstest.TheOneRingOracle)
	if !ok {
		t.Fatalf("The One Ring is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("every clause of The One Ring is implemented since #1197: %+v", spec)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushPermanentForTest(g, me.ID, "The One Ring", effectstest.TheOneRingOracle, "Legendary Artifact")
	c, ok := g.LookupCardForEffect(ring)
	if !ok {
		t.Fatalf("The One Ring is not on the battlefield")
	}
	if !game.HasKeyword(&c, "indestructible") {
		t.Errorf("The One Ring should be indestructible: %+v", c.Effective().Abilities)
	}
}
