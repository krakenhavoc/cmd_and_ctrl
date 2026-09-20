package effects

import "testing"

const quicksmithGeniusOracle = "bca0c136-3845-4cb9-8d80-f504a01ad50a"

// TestQuicksmithGeniusDrawsOnlyAfterTheDiscardIsAnswered pins the
// ordering the older DiscardChoiceForEffect used to get wrong: the
// discard prompt settles first, and only THEN does the draw happen.
func TestQuicksmithGeniusDrawsOnlyAfterTheDiscardIsAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Quicksmith Genius", "Creature — Human Artificer", quicksmithGeniusOracle, nil)
	passPriorityAroundTable(t, g)

	hand, library := me.Hand.Size(), me.Library.Size()
	castWithCost(t, g, "Signet", "Artifact", "{2}", "")
	passPriorityAroundTable(t, g)

	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the trigger offers up to one discard: owes %d", got)
	}
	if me.Library.Size() != library {
		t.Fatalf("no draw before the discard is answered: library %d, want %d", me.Library.Size(), library)
	}
	// Casting the Signet added it to hand and immediately cast it
	// away again, so the hand is back to its pre-cast size until the
	// discard prompt is answered.
	if me.Hand.Size() != hand {
		t.Fatalf("no discard has happened yet: hand %d, want %d", me.Hand.Size(), hand)
	}

	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)

	// Discard one, draw one: the hand SIZE is net unchanged (a
	// different card), but the library is down one and the graveyard
	// is up one — the observable proof the draw actually happened
	// after the discard rather than being skipped.
	if me.Hand.Size() != hand {
		t.Errorf("discard one, draw one: net hand size should be %d, got %d", hand, me.Hand.Size())
	}
	if me.Library.Size() != library-1 {
		t.Errorf("the draw only happens after the discard is chosen: library %d, want %d", me.Library.Size(), library-1)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("the discarded card should be in the graveyard: graveyard size %d, want 1", me.Graveyard.Size())
	}
}

// TestQuicksmithGeniusDecliningDrawsNothing: "if you do" is a real
// condition, and declining is legal.
func TestQuicksmithGeniusDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Quicksmith Genius", "Creature — Human Artificer", quicksmithGeniusOracle, nil)
	passPriorityAroundTable(t, g)

	hand, library := me.Hand.Size(), me.Library.Size()
	castWithCost(t, g, "Signet", "Artifact", "{2}", "")
	passPriorityAroundTable(t, g)

	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand {
		t.Errorf("declining discards nothing: hand %d, want %d", me.Hand.Size(), hand)
	}
	if me.Library.Size() != library {
		t.Errorf("declining draws nothing: library %d, want %d", me.Library.Size(), library)
	}
}
