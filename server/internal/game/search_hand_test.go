package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_hand_test.go — #3988: "search that player's graveyard, hand,
// and library" (SearchLibrarySpec.AlsoGraveyardAndHand), the extraction
// search Kotose, the Silent Spider prints. Its card test is in
// cards/effects.

func extractionHandCard(p *Player, name string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	return id
}

func extractionSpec(searcher, victim uuid.UUID) SearchLibrarySpec {
	return SearchLibrarySpec{
		Player: searcher, LibraryOwner: victim, Pred: named("Wanted"), Dest: ZoneExile,
		Unbounded: true, Shuffle: true, AlsoGraveyardAndHand: true,
	}
}

// The three zones in one prompt, each candidate marked by zone; the
// searcher sees the whole hand; any number, none included.
func TestExtractionSearchOffersAllThreeZones(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	lib := searchLibraryOf(them, "Wanted", "Other")
	gy := gyCard(them, "Wanted")
	hand := extractionHandCard(them, "Wanted")
	other := extractionHandCard(them, "Other")
	runSearch(t, g, extractionSpec(me.ID, them.ID))

	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if len(c.SearchCards) != 3 || c.SearchMax != 3 || c.SearchMin != 0 {
		t.Fatalf("candidates %v max %d min %d, want three, any number", c.SearchCards, c.SearchMax, c.SearchMin)
	}
	if len(c.SearchGraveyardCards) != 1 || c.SearchGraveyardCards[0] != gy {
		t.Errorf("graveyard half %v", c.SearchGraveyardCards)
	}
	if len(c.SearchHandCards) != 1 || c.SearchHandCards[0] != hand {
		t.Errorf("hand half %v", c.SearchHandCards)
	}
	if !c.SearchesGraveyardHandAndLibrary() {
		t.Error("the prompt does not say it searches all three zones")
	}
	for _, id := range []uuid.UUID{hand, other} {
		if card, _ := g.LookupCardForEffect(id); !card.IsKnownTo(me.ID) {
			t.Errorf("hand card %s not seen by the searcher", id)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lib[0], gy, hand}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{lib[0], gy, hand} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !them.Hand.Contains(other) {
		t.Error("the card of another name left the hand")
	}
}

// All three zones are named, so a take wholly out of the graveyard and
// hand still searched the library: it shuffles and fires the event, unlike
// "and/or graveyard".
func TestExtractionSearchAlwaysSearchesTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	searchLibraryOf(them, "Other", "Other", "Other")
	knowLibrary(them)
	gy := gyCard(them, "Wanted")
	hand := extractionHandCard(them, "Wanted")
	runSearch(t, g, extractionSpec(me.ID, them.ID))
	c := openSearch(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gy, hand}); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 {
		t.Errorf("%d search events, want 1", searchEvents(g))
	}
	if libraryKnownTo(them, them.ID) != 0 {
		t.Error("the library was not shuffled")
	}
}

// A hand candidate is checked against the hand it was offered from: one
// that has left it since is refused, and does not count toward the floor.
func TestExtractionSearchRefusesAHandCardThatLeft(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	searchLibraryOf(them, "Other")
	hand := extractionHandCard(them, "Wanted")
	runSearch(t, g, extractionSpec(me.ID, them.ID))
	c := openSearch(g, me.ID)
	g.WithWriteLock(func() {
		card, err := them.Hand.Remove(hand)
		if err != nil {
			t.Fatal(err)
		}
		them.Graveyard.PushTop(card)
	})
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{hand}); err == nil {
		t.Error("a hand candidate that left the hand was taken")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Errorf("finding nothing: %v", err)
	}
}

func TestCloneCarriesTheHandCandidates(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	searchLibraryOf(them, "Other")
	hand := extractionHandCard(them, "Wanted")
	runSearch(t, g, extractionSpec(me.ID, them.ID))
	cl := g.Clone()
	c := openSearch(cl, me.ID)
	if c == nil || len(c.SearchHandCards) != 1 || c.SearchHandCards[0] != hand {
		t.Fatalf("clone lost the hand half: %+v", c)
	}
	c.SearchHandCards[0] = uuid.New()
	if openSearch(g, me.ID).SearchHandCards[0] != hand {
		t.Error("the clone shares its backing array with the live prompt")
	}
}
