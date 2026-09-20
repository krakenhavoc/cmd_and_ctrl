package effects

import (
	"testing"

	"github.com/google/uuid"
)

const captainHowlerOracle = "46c99d67-8307-4af0-afd0-377714806614"

// TestCaptainHowlerSeaScourgeBatchesToOneTriggerWithTheSingleCardAmount
// pins the declared caveat: discarding two cards at once is
// OncePerBatch'd to a single trigger (one target prompt, not two —
// see the "batch size" engine seam), and the pump it applies is the
// single-card +2/+0 rather than the printed +4/+0.
func TestCaptainHowlerSeaScourgeBatchesToOneTriggerWithTheSingleCardAmount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	howler := pushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	if me.Hand.Size() < 2 {
		t.Fatalf("need two cards in hand to discard: have %d", me.Hand.Size())
	}
	discarded := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 2) })
	answerDiscard(t, g, me.ID, discarded...)

	pickCard(t, g, me.ID, bear)
	if got := triggersOnStackFrom(g, howler); got != 1 {
		t.Fatalf("a two-card discard must fire the trigger once, not once per card: got %d", got)
	}

	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("the declared caveat: a two-card discard still pumps by +2/+0, not +2/+0 per card: power %d, want 4", p)
	}
}

// TestCaptainHowlerSeaScourgeOneCardIsExact confirms the ordinary,
// non-caveated case: discarding exactly one card gets the exact
// printed pump.
func TestCaptainHowlerSeaScourgeOneCardIsExact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	discardFromHand(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("one discarded card is +2/+0 exactly: power %d, want 4", p)
	}
}
