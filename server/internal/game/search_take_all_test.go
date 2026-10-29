package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_take_all_test.go — #4033: "search that player's graveyard, hand,
// and library for ALL cards with the same name" (SearchLibrarySpec.TakeAll),
// which Eradicate, Extirpate, Quash, Scour, Splinter and Sowing Salt print.
// The graveyard is public, so its matches are found; the hand and the
// library are hidden, so the searcher may leave theirs (CR 701.23b).

func takeAllSpec(searcher, victim uuid.UUID) SearchLibrarySpec {
	spec := extractionSpec(searcher, victim)
	spec.Unbounded = false
	spec.TakeAll = true
	return spec
}

// The graveyard's matches are not offered and are exiled whatever the
// searcher answers; the hidden zones' matches are offered, any number.
func TestTakeAllFindsEveryGraveyardMatch(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	lib := searchLibraryOf(them, "Wanted", "Other")
	gy1, gy2 := gyCard(them, "Wanted"), gyCard(them, "Wanted")
	gyOther := gyCard(them, "Other")
	hand := extractionHandCard(them, "Wanted")
	var then []uuid.UUID
	spec := takeAllSpec(me.ID, them.ID)
	spec.ThenTail = searchHook(func(_ *Game, found []uuid.UUID) error { then = found; return nil })
	runSearch(t, g, spec)

	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the hidden matches")
	}
	if len(c.SearchCards) != 2 || len(c.SearchGraveyardCards) != 0 {
		t.Fatalf("candidates %v (graveyard %v), want the hand and library matches only", c.SearchCards, c.SearchGraveyardCards)
	}
	if c.SearchMax != 2 || g.SearchFloorLocked(c) != 0 {
		t.Errorf("max %d floor %d, want 2 and 0: a hidden zone may come up short", c.SearchMax, g.SearchFloorLocked(c))
	}
	if !g.SearchPickLegalLocked(c, nil) {
		t.Error("finding nothing in the hidden zones was refused")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{hand}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{gy1, gy2, hand} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !them.Library.Contains(lib[0]) || !them.Graveyard.Contains(gyOther) {
		t.Error("a card nobody found was exiled")
	}
	if len(then) != 3 {
		t.Errorf("Then got %v, want the two graveyard cards and the hand card", then)
	}
	if searchEvents(g) != 1 {
		t.Errorf("%d search events, want 1", searchEvents(g))
	}
}

// With nothing in the hidden zones there is no choice: the graveyard's
// matches are exiled without a prompt, and the library is still searched
// and shuffled.
func TestTakeAllWithOnlyGraveyardMatchesDoesNotAsk(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	searchLibraryOf(them, "Other", "Other")
	knowLibrary(them)
	gy := gyCard(them, "Wanted")
	runSearch(t, g, takeAllSpec(me.ID, them.ID))
	if c := openSearch(g, me.ID); c != nil {
		t.Fatalf("asked with nothing to choose: %+v", c)
	}
	if !g.Exile.Contains(gy) {
		t.Error("the graveyard match was not exiled")
	}
	if searchEvents(g) != 1 || libraryKnownTo(them, them.ID) != 0 {
		t.Error("the library was not searched and shuffled")
	}
}

// A graveyard match that left the graveyard while the prompt was open is
// not there to find.
func TestTakeAllSkipsAGraveyardMatchThatLeft(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	them.Hand.Cards = nil
	searchLibraryOf(them, "Wanted")
	gy := gyCard(them, "Wanted")
	runSearch(t, g, takeAllSpec(me.ID, them.ID))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	g.WithWriteLock(func() {
		card, err := them.Graveyard.Remove(gy)
		if err != nil {
			t.Fatal(err)
		}
		them.Hand.PushTop(card)
	})
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !them.Hand.Contains(gy) {
		t.Error("a card that left the graveyard was exiled from where it went")
	}
}

// "And/or graveyard" lets the searcher skip the graveyard, so TakeAll
// forces nothing there.
func TestTakeAllNeedsTheGraveyardHandAndLibrarySearch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Unbounded: true,
		Shuffle: true, AlsoGraveyard: true, TakeAll: true,
	})
	c := openSearch(g, me.ID)
	if c == nil || len(c.SearchGraveyardCards) != 1 || c.SearchGraveyardCards[0] != gy {
		t.Fatalf("the and/or graveyard match is not a candidate: %+v", c)
	}
	if !g.SearchPickLegalLocked(c, nil) {
		t.Error("finding nothing was refused")
	}
}
