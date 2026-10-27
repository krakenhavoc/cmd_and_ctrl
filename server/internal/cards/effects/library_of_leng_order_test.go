package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_order_test.go — #2949, ADR 0088 amendment 2026-10-14:
// the cards of one discard that Library of Leng puts on top of the
// library are put there at the same time, so their owner orders them
// (CR 401.4), and the rest of the effect waits for that order.

// lengBoard is Library of Leng under seat 0, an empty hand, a library of
// filler, and n cards in hand to discard.
func lengBoard(t *testing.T, n int) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	awbPadLibrary(me)
	var cards []uuid.UUID
	for i := 0; i < n; i++ {
		cards = append(cards, handCardFull(me, "Pitch", "Sorcery", "{1}", "", nil))
	}
	return g, me, cards
}

// lengDiscardThen is lengDiscard with a "then" that counts its runs.
func lengDiscardThen(t *testing.T, g *game.Game, who *game.Player, ran *int, cards ...uuid.UUID) {
	t.Helper()
	var promptID uuid.UUID
	// #4361 S7b: the "then" is a keyed discard-then.
	then := game.RegisterDiscardThen("test/4361-s7b-leng-"+uuid.NewString(), func(*game.Game, game.DiscardsDone) error { *ran++; return nil })
	g.WithWriteLock(func() {
		if err := g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{Player: who.ID, N: len(cards)}, then, nil); err != nil {
			t.Fatalf("PlayerDiscardsThenKeyedForEffect: %v", err)
		}
		if n := len(g.PendingChoices); n > 0 {
			promptID = g.PendingChoices[n-1].ID
		}
	})
	if err := g.ResolveChooseCards(promptID, who.ID, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func putInLibraryPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary {
			return c
		}
	}
	return nil
}

func lengLibraryTop(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := len(p.Library.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

// Two and three saved cards: the owner is asked for the order, the
// chosen (non-default) order is the library's, and the "then" waits.
func TestLibraryOfLengOwnerOrdersSeveralSavedCards(t *testing.T) {
	for _, n := range []int{2, 3} {
		g, me, cards := lengBoard(t, n)
		ran := 0
		lengDiscardThen(t, g, me, &ran, cards...)
		for i := 0; i < n; i++ {
			answerLengPrompt(t, g, me.ID, true)
		}
		prompt := putInLibraryPrompt(g)
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
		if got := lengLibraryTop(me, n); !sameIDs(got, want) {
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
	g, me, cards := lengBoard(t, 2)
	ran := 0
	lengDiscardThen(t, g, me, &ran, cards...)
	answerLengPrompt(t, g, me.ID, true)
	answerLengPrompt(t, g, me.ID, false)
	if putInLibraryPrompt(g) != nil {
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
	g, me, cards := lengBoard(t, 2)
	opp := g.Seats[1]
	w := &lengWatcher{}
	g.RegisterListener(w)
	ran := 0
	lengDiscardThen(t, g, me, &ran, cards...)
	answerLengPrompt(t, g, me.ID, true)
	answerLengPrompt(t, g, me.ID, true)
	prompt := putInLibraryPrompt(g)
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
