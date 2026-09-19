package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// inapp_trigger_reports_test.go — the reproduction behind the in-app
// "my trigger didn't fire" reports (#366, #369, #370, #373, #508).
//
// Four of the five named a card with no Spec, so there was no trigger
// to fire at all; those are pinned in the game package, where
// TestNeedsCatalogEffectCatchesRealRules asserts the engine at least
// SAYS SO to the player who cast one. The fifth, #370, named a card
// that does have a Spec, and that is this file.

// TestIssue370MaryReadVehicleDiscardMakesATappedTreasure walks the arm
// of Mary Read and Anne Bonny the reporter used and no test covered:
// tap to loot, then pitch a Vehicle to the prompt.
//
// The three arms that were already covered each reach the discard a
// different way — DiscardRandomForEffect for the Island
// (TestMaryReadLootsAndTurnsTypedDiscardsIntoTreasure), an additional
// cost for the Pirate (TestAdditionalCostDiscardTriggersTheCommander),
// the CR 514.1 hand-size discard for the cleanup step
// (TestCleanupDiscardWatchersFireInTheDiscardingTurn). The one
// route nothing took was the one a player takes.
//
// The assertion that matters is the FIRST one, and it is the one that
// used to fail. On the 2026-09-11 build the report came from, the
// prompt was answered by the `discard_selection` action against
// Game.DiscardPending: it moved the card and emitted EventDiscardCard,
// then returned without running state checks. The Treasure trigger sat
// on PendingTriggers — invisible: not on the board, not on the stack
// overlay, nothing in the snapshot the reporter was looking at when
// they filed. It reached the stack only once somebody next passed
// priority, which is a different turn's worth of clicks away.
//
// CR 117.5 puts triggered abilities on the stack the next time a
// player would receive priority, and answering a discard that is part
// of a resolving ability (CR 608.2c) is exactly that moment. #797 made
// an effect discard a PendingChoice whose answer path ends in
// runStateChecksLocked, so the trigger is announced by the same action
// that answers the prompt.
func TestIssue370MaryReadVehicleDiscardMakesATappedTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mary := pushCatalogPermanent(g, me.ID, "Mary Read and Anne Bonny",
		"Legendary Creature — Human Assassin Pirate", maryReadOracle, false)

	// Empty the hand so the Vehicle is the only discardable card.
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			c, _ := me.Hand.Top()
			_, _ = game.MoveCard(me.Hand, me.Library, c.InstanceID)
		}
	})
	vehicle := handCard(me, "Smuggler's Copter", "Artifact — Vehicle")

	if err := g.ActivateCatalogAbility(me.ID, mary, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Mary Read's loot: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the loot should owe one discard, owed %d", got)
	}
	answerDiscard(t, g, me.ID, vehicle)

	// Announced by the same action that answered the prompt, so the
	// player who pitched the Vehicle can see what it bought them.
	if triggerOnStack(g, mary) == nil {
		t.Fatalf("the Treasure trigger should be on the stack as the discard is answered")
	}

	passPriorityAroundTable(t, g)

	// Exactly one, and tapped — the printed card, and the half of the
	// report that #530 had made unreadable at the table.
	count := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name != "Treasure" {
			continue
		}
		count++
		if !c.Tapped {
			t.Errorf("Mary Read's Treasure should enter tapped")
		}
	}
	if count != 1 {
		t.Fatalf("discarding a Vehicle should make exactly one Treasure, found %d", count)
	}
}
