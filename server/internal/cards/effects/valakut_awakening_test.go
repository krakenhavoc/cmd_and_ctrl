package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const valakutAwakeningOracle = "ff0ab867-b710-4b1a-baed-95fc3cf68f79"

// TestValakutAwakeningPutsCardsOnBottomThenDrawsThatManyPlusOne pins
// the whole card: the pile chosen goes to the BOTTOM of the library
// (not the graveyard, not the top), and the draw is exactly one more
// than the pile.
func TestValakutAwakeningPutsCardsOnBottomThenDrawsThatManyPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Library Card")
	excess1 := pushHandCardWithManaCost(me, "Excess One", "Instant", "{1}")
	excess2 := pushHandCardWithManaCost(me, "Excess Two", "Instant", "{1}")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Valakut Awakening", "Instant", valakutAwakeningOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt over the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{excess1, excess2}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no put_in_library prompt for the bottom order")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{excess1, excess2}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Contains(excess1) || me.Hand.Contains(excess2) {
		t.Error("the put-back cards are still in hand")
	}
	if bottom, err := me.Library.Bottom(); err != nil || (bottom.InstanceID != excess1 && bottom.InstanceID != excess2) {
		t.Errorf("the put-back cards did not land on the bottom of the library: %+v, err %v", bottom, err)
	}
	if got := me.Hand.Size(); got != handBefore-2+3 {
		t.Errorf("hand size %d, want %d (put back 2, draw 3)", got, handBefore-2+3)
	}
}

// Declining to put anything back still draws one card.
func TestValakutAwakeningDrawsOneWithNothingPutBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Library Card")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Valakut Awakening", "Instant", valakutAwakeningOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt over the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d, want %d (draw 1, nothing put back)", got, handBefore+1)
	}
}
