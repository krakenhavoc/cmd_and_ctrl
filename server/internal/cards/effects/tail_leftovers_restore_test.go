package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tail_leftovers_restore_test.go — ADR 0041 P9, the tier 4 tail's
// design leftovers (#1497). Each card here used to keep a trigger-time
// value in its Build's closure; now the value is data the snapshot
// carries, and these restore the table with the trigger waiting and
// resolve it in the restored game.

// waitingTriggerFor is the trigger `cardID` put up, queued or on the
// stack.
func waitingTriggerFor(g *game.Game, cardID uuid.UUID) *game.StackItem {
	if it := pendingTriggerFor(g, cardID); it != nil {
		return it
	}
	return triggerOnStack(g, cardID)
}

// The Ozolith's "those counters" is read at resolution off the #1379
// record, which the snapshot carries: a restored trigger still banks
// what the creature had.
func TestRestoredOzolithTriggerStillCollects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oz := seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)
	creature := seedCreatureWithCounters(g, me.ID, map[string]int{"+1/+1": 3, "stun": 1})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	corpusRequireTriggeredStamp(t, waitingTriggerFor(g, oz), "own:")

	restored := restoreThroughJSON(t, g)
	passPriorityAroundTable(t, restored)
	got := allCountersOn(t, restored, oz)
	if got["+1/+1"] != 3 || got["stun"] != 1 {
		t.Errorf("the restored Ozolith's counters = %v, want +1/+1:3 stun:1", got)
	}
}

// Valakut Exploration's "cards exiled with this enchantment" rides the
// item's Payload: a restored end-step trigger sweeps the same cards.
func TestRestoredValakutExplorationSweepStillSweeps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	valakut := seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)
	loot := stackLibrary(me, "Stashed Bolt", "Instant", "{R}")
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatal("setup: landfall did not exile the top card")
	}
	advanceTo(t, g, game.StepEnd)
	it := waitingTriggerFor(g, valakut)
	corpusRequireTriggeredStamp(t, it, "own:")
	if len(it.Payload) != 1 || it.Payload[0].ID != loot {
		t.Fatalf("the sweep's payload = %v, want the exiled card", it.Payload)
	}

	restored := restoreThroughJSON(t, g)
	rMe, rOpp := restored.Seats[0], restored.Seats[1]
	before := rOpp.Life
	passPriorityAroundTable(t, restored)
	if !rMe.Graveyard.Contains(loot) {
		t.Error("the restored sweep did not put the exiled card into the graveyard")
	}
	if got := before - rOpp.Life; got != 1 {
		t.Errorf("opponent took %d damage after the restore, want 1", got)
	}
}

// Molten Primordial's clauses are rebuilt on restore from the seat list
// and the controller the trigger context recorded — not the source's
// controller now. Stolen while its trigger waits, the restored
// Primordial still takes one creature from each of ITS opponents.
func TestRestoredMoltenPrimordialKeepsItsOpponentsAfterASteal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	b := pushCreatureToBattlefieldForTest(g, opps[1].ID, "B's Bear")
	primordial := castHoldCreature(t, g, "Molten Primordial", "Creature — Avatar", oracleMoltenPrimordial, 6, 4)
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	corpusRequireTriggeredStamp(t, waitingTriggerFor(g, primordial), "own:")
	stealForTest(t, g, primordial, opps[0].ID)

	restored := restoreThroughJSON(t, g)
	passPriorityAroundTable(t, restored)
	for _, id := range []uuid.UUID{a, b} {
		if c, ok := battlefieldCard(restored, id); !ok || c.Controller != me.ID {
			t.Errorf("%s: controller %v, want the Primordial's controller as it entered", id, c.Controller)
		}
	}
}
