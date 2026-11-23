package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Two and three saved cards: the owner is asked for the order, the
// chosen (non-default) order is the library's, and the "then" waits.
func TestLibraryOfLengOwnerOrdersSeveralSavedCards(t *testing.T) {
	for _, n := range []int{2, 3} {
		g, me, cards := effectstest.LengBoard(t, n)
		ran := 0
		effectstest.LengDiscardThen(t, g, me, &ran, cards...)
		for i := 0; i < n; i++ {
			effectstest.AnswerLengPrompt(t, g, me.ID, true)
		}
		prompt := effectstest.PutInLibraryPrompt(g)
		if prompt == nil {
			t.Fatalf("%d saved cards: no ordering prompt", n)
		}
		if prompt.Chooser != me.ID || prompt.LibraryPlacement != game.LibraryPlaceTop {
			t.Errorf("prompt chooser %v placement %q, want the owner and top", prompt.Chooser, prompt.LibraryPlacement)
		}
		if ran != 0 {
			t.Fatalf("%d saved cards: the effect's then ran before the order was chosen", n)
		}
		// The reverse of the default: the first card discarded on top.
		want := append([]uuid.UUID(nil), cards...)
		if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		if got := effectstest.LengLibraryTop(me, n); !effectstest.SameIDs(got, want) {
			t.Errorf("%d saved cards: library top %v, want the chosen %v", n, got, want)
		}
		if ran != 1 {
			t.Errorf("%d saved cards: then ran %d times, want once after the order", n, ran)
		}
	}
}

// One card saved and one declined, or only one card discarded: nothing
// to order, and the then runs at once.
func TestLibraryOfLengOneSavedCardIsNotOrdered(t *testing.T) {
	g, me, cards := effectstest.LengBoard(t, 2)
	ran := 0
	effectstest.LengDiscardThen(t, g, me, &ran, cards...)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	effectstest.AnswerLengPrompt(t, g, me.ID, false)
	if effectstest.PutInLibraryPrompt(g) != nil {
		t.Fatal("one saved card is not an order to choose")
	}
	if ran != 1 {
		t.Errorf("then ran %d times, want once", ran)
	}
	if !me.Library.Contains(cards[0]) || !me.Graveyard.Contains(cards[1]) {
		t.Error("the saved card is in the library and the declined one in the graveyard")
	}
}

// Every discard still fires its event, and opponents learn nothing of
// the order: only the owner knows the cards after it is chosen.
func TestLibraryOfLengOrderKeepsDiscardEventsAndPrivacy(t *testing.T) {
	g, me, cards := effectstest.LengBoard(t, 2)
	opp := g.Seats[1]
	w := &lengWatcher{}
	g.RegisterListener(w)
	ran := 0
	effectstest.LengDiscardThen(t, g, me, &ran, cards...)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	prompt := effectstest.PutInLibraryPrompt(g)
	if prompt == nil {
		t.Fatal("no ordering prompt")
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, cards); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if len(w.discards) != 2 {
		t.Errorf("EventDiscardCard x %d, want 2", len(w.discards))
	}
	for _, id := range cards {
		for i := range me.Library.Cards {
			if c := &me.Library.Cards[i]; c.InstanceID == id && c.IsKnownTo(opp.ID) {
				t.Errorf("an opponent knows a library card after the order (%v)", id)
			}
		}
	}
}
