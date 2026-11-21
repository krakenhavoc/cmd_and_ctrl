package testsav

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The loot ability draws and prompts, and the discard trigger exiles
// what was pitched. The three abilities share one {T}, so the cash-out
// waits for another turn.
func TestBagOfHoldingLootsThenExilesWhatWasPitched(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bag := effectstest.B12Push(g, me.ID, "Bag of Holding", "Artifact", effectstest.BagOfHoldingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyHandToLibrary(g, me)
	pitched := effectstest.HandCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	if err := g.ActivateCatalogAbility(me.ID, bag, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the loot ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Fatalf("hand %d after the draw, want 2 (the pitched card plus the draw)", me.Hand.Size())
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("the loot ability owes a discard of 1, got %d", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitched)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(pitched) {
		t.Fatal("the discarded card is exiled from the graveyard")
	}
	if me.Graveyard.Contains(pitched) {
		t.Error("the discarded card does not stay in the graveyard")
	}
	// All three abilities share the one {T}, so the Bag cannot loot
	// and cash out on the same turn.
	if err := g.ActivateCatalogAbility(me.ID, bag, 1, game.ActivateAbilityParams{}); err != game.ErrAlreadyTapped {
		t.Errorf("the tapped Bag: got %v, want ErrAlreadyTapped", err)
	}
}

// Only what this Bag exiled comes back. A card that reached exile by
// some other route was never "exiled with this artifact" and stays
// where it is — the record is keyed on the Bag and on the trigger's
// label, not on "everything in exile".
func TestBagOfHoldingReturnsOnlyTheCardsItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bag := effectstest.B12Push(g, me.ID, "Bag of Holding", "Artifact", effectstest.BagOfHoldingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyHandToLibrary(g, me)
	stranger := effectstest.PushGraveyardCardTyped(me, "Stranger", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(stranger) })

	pitched := effectstest.HandCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatal("the discard trigger exiles the pitched card")
	}

	effectstest.B16Activate(t, g, me.ID, bag, 1, game.ActivateAbilityParams{})
	if !me.Hand.Contains(pitched) {
		t.Error("the Bag's own exile comes back")
	}
	if !g.Exile.Contains(stranger) {
		t.Error("a card exiled by something else stays in exile")
	}
}

// An opponent's discard is not yours, and a card that is no longer in
// a graveyard when the trigger resolves is not exiled.
func TestBagOfHoldingOnlyWatchesItsControllerAndOnlyTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Bag of Holding", "Artifact", effectstest.BagOfHoldingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyHandToLibrary(g, opp)
	theirs := effectstest.HandCardFull(opp, "Their Card", "Instant", "{R}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Error("an opponent's discard is not yours")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("their card stays in their graveyard")
	}

	effectstest.EmptyHandToLibrary(g, me)
	mine := effectstest.HandCardFull(me, "Mine", "Instant", "{U}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	// Answered in response: the card leaves the graveyard before the
	// trigger resolves, so there is nothing to exile from there.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(mine) {
		t.Error("a card that left the graveyard first is not exiled")
	}
	if !me.Hand.Contains(mine) {
		t.Error("the card that was rescued stays rescued")
	}
}
