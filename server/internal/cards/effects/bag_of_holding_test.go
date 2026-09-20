package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bagOfHoldingOracle = "63c04040-e109-494e-baa6-c639a6c9a996"

// The loot ability draws and prompts, and the discard trigger exiles
// what was pitched. The three abilities share one {T}, so the cash-out
// waits for another turn.
func TestBagOfHoldingLootsThenExilesWhatWasPitched(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := b12Push(g, me.ID, "Bag of Holding", "Artifact", bagOfHoldingOracle, 0, 0)
	advanceToMain(t, g)
	emptyHandToLibrary(g, me)
	pitched := handCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	if err := g.ActivateCatalogAbility(me.ID, bag, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the loot ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Fatalf("hand %d after the draw, want 2 (the pitched card plus the draw)", me.Hand.Size())
	}
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("the loot ability owes a discard of 1, got %d", discardOwed(g, me.ID))
	}
	answerDiscard(t, g, me.ID, pitched)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := b12Push(g, me.ID, "Bag of Holding", "Artifact", bagOfHoldingOracle, 0, 0)
	advanceToMain(t, g)
	emptyHandToLibrary(g, me)
	stranger := pushGraveyardCardTyped(me, "Stranger", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(stranger) })

	pitched := handCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatal("the discard trigger exiles the pitched card")
	}

	b16Activate(t, g, me.ID, bag, 1, game.ActivateAbilityParams{})
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Bag of Holding", "Artifact", bagOfHoldingOracle, 0, 0)
	advanceToMain(t, g)
	emptyHandToLibrary(g, opp)
	theirs := handCardFull(opp, "Their Card", "Instant", "{R}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Error("an opponent's discard is not yours")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("their card stays in their graveyard")
	}

	emptyHandToLibrary(g, me)
	mine := handCardFull(me, "Mine", "Instant", "{U}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	// Answered in response: the card leaves the graveyard before the
	// trigger resolves, so there is nothing to exile from there.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(mine) })
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(mine) {
		t.Error("a card that left the graveyard first is not exiled")
	}
	if !me.Hand.Contains(mine) {
		t.Error("the card that was rescued stays rescued")
	}
}
