package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_zones_test.go — #4516: "search your library and/or graveyard"
// records which zones the searcher looked in. The printed text shuffles
// only "if you search your library this way" (CR 701.23), so a search of
// the graveyard alone, found nothing or not, does not shuffle; a library
// pick is a library search; and a player who took a graveyard card, or
// nothing, can say they searched the library too.

// zonesSearch opens an "and/or graveyard" search for "Wanted" with one
// match in each zone (the library's known to the seat, so a shuffle is
// observable as un-knowing it) and returns the prompt and both ids.
func zonesSearch(t *testing.T, g *Game, spec SearchLibrarySpec) (c *PendingChoice, lib, gy uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	lib = searchLibraryOf(me, "Wanted", "Other")[0]
	knowLibrary(me)
	gy = gyCard(me, "Wanted")
	spec.Player = me.ID
	spec.Pred = named("Wanted")
	if spec.Dest == "" {
		spec.Dest = ZoneHand
	}
	spec.Limit = 1
	spec.Shuffle = true
	spec.AlsoGraveyard = true
	runSearch(t, g, spec)
	c = openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	return c, lib, gy
}

// Nothing in either zone matches and the searcher found nothing: they
// searched the graveyard alone, so the library keeps its order and no
// search is announced. The card's "then" still runs.
func TestGraveyardOnlySearchFindingNothingDoesNotShuffle(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other", "Another")
	knowLibrary(me)
	var ran bool
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
		ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { ran = len(found) == 0; return nil }),
	})
	c := openSearch(g, me.ID)
	if c == nil || len(c.SearchCards) != 0 {
		t.Fatalf("an empty search still asks: %+v", c)
	}
	if !c.SearchAsksLibraryZone() {
		t.Error("an \"and/or graveyard\" prompt must ask which zones were searched")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 0 || libraryKnownTo(me, me.ID) != 2 {
		t.Errorf("events=%d known=%d: a graveyard-only search that found nothing must not shuffle",
			searchEvents(g), libraryKnownTo(me, me.ID))
	}
	if !ran {
		t.Error("the search's \"then\" did not run with nothing found")
	}
}

// The same empty search, when the searcher says they looked in the
// library too: that is a library search, so it shuffles and is announced.
func TestFindingNothingAfterSearchingTheLibraryShuffles(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other", "Another")
	knowLibrary(me)
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if err := g.ResolveSearchLibraryZones(c.ID, me.ID, nil, true); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: a library search that found nothing shuffles",
			searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// A pick from the graveyard is a graveyard-only search unless the
// searcher says the library was searched too.
func TestGraveyardPickShufflesOnlyIfTheLibraryWasSearched(t *testing.T) {
	for _, tc := range []struct {
		name       string
		libraryToo bool
		wantEvents int
		wantKnown  int
	}{
		{"graveyard only", false, 0, 2},
		{"library searched too", true, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			c, _, gy := zonesSearch(t, g, SearchLibrarySpec{})
			if err := g.ResolveSearchLibraryZones(c.ID, me.ID, []uuid.UUID{gy}, tc.libraryToo); err != nil {
				t.Fatal(err)
			}
			if !me.Hand.Contains(gy) {
				t.Fatal("the graveyard card did not reach the hand")
			}
			if searchEvents(g) != tc.wantEvents || libraryKnownTo(me, me.ID) != tc.wantKnown {
				t.Errorf("events=%d known=%d, want %d/%d", searchEvents(g), libraryKnownTo(me, me.ID),
					tc.wantEvents, tc.wantKnown)
			}
		})
	}
}

// A library pick is a library search, whatever the answer says about it.
func TestLibraryPickShufflesWhateverTheAnswerSays(t *testing.T) {
	for _, libraryToo := range []bool{false, true} {
		g := newActiveGame(t)
		me := g.Seats[0]
		c, lib, _ := zonesSearch(t, g, SearchLibrarySpec{})
		if err := g.ResolveSearchLibraryZones(c.ID, me.ID, []uuid.UUID{lib}, libraryToo); err != nil {
			t.Fatal(err)
		}
		if !me.Hand.Contains(lib) {
			t.Fatal("the library card did not reach the hand")
		}
		if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
			t.Errorf("libraryToo=%v: events=%d known=%d, want a shuffle", libraryToo,
				searchEvents(g), libraryKnownTo(me, me.ID))
		}
	}
}

// A graveyard card onto the battlefield finishes through the entry
// pipeline's tail; the library-searched answer rides it.
func TestGraveyardPickToTheBattlefieldKeepsTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		libraryToo bool
		wantEvents int
	}{{false, 0}, {true, 1}} {
		g := newActiveGame(t)
		me := g.Seats[0]
		c, _, gy := zonesSearch(t, g, SearchLibrarySpec{Dest: ZoneBattlefield})
		if err := g.ResolveSearchLibraryZones(c.ID, me.ID, []uuid.UUID{gy}, tc.libraryToo); err != nil {
			t.Fatal(err)
		}
		if !g.Battlefield.Contains(gy) {
			t.Fatal("the graveyard card did not enter the battlefield")
		}
		if searchEvents(g) != tc.wantEvents {
			t.Errorf("libraryToo=%v: events=%d, want %d", tc.libraryToo, searchEvents(g), tc.wantEvents)
		}
	}
}

// The answer means nothing to a search with no graveyard half: a plain
// library search that finds nothing shuffles, and so does "graveyard,
// hand, and library", which names the library.
func TestOnlyAndOrGraveyardSearchesAskWhichZones(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	knowLibrary(me)
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true})
	c := openSearch(g, me.ID)
	if c.SearchAsksLibraryZone() {
		t.Error("a library-only search has no zone to ask about")
	}
	if err := g.ResolveSearchLibraryZones(c.ID, me.ID, nil, false); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: a library search shuffles", searchEvents(g), libraryKnownTo(me, me.ID))
	}

	g = newActiveGame(t)
	me = g.Seats[0]
	searchLibraryOf(me, "Other")
	knowLibrary(me)
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyardAndHand: true,
	})
	c = openSearch(g, me.ID)
	if c.SearchAsksLibraryZone() {
		t.Error("\"graveyard, hand, and library\" names every zone, so there is nothing to ask")
	}
	if err := g.ResolveSearchLibraryZones(c.ID, me.ID, nil, false); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: the all-zones search shuffles", searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// A restore while the prompt is open keeps the question, and the answer
// given on the restored table decides the shuffle the same way.
func TestARestoredAndOrGraveyardSearchStillAsksWhichZones(t *testing.T) {
	for _, tc := range []struct {
		name       string
		picks      func(lib, gy uuid.UUID) []uuid.UUID
		libraryToo bool
		wantEvents int
	}{
		{"nothing, graveyard only", func(_, _ uuid.UUID) []uuid.UUID { return nil }, false, 0},
		{"nothing, library too", func(_, _ uuid.UUID) []uuid.UUID { return nil }, true, 1},
		{"graveyard card, graveyard only", func(_, gy uuid.UUID) []uuid.UUID { return []uuid.UUID{gy} }, false, 0},
		{"graveyard card, library too", func(_, gy uuid.UUID) []uuid.UUID { return []uuid.UUID{gy} }, true, 1},
		{"library card", func(lib, _ uuid.UUID) []uuid.UUID { return []uuid.UUID{lib} }, false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			_, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
			g = strictSearchRestore(t, g)
			me := g.Seats[0]
			knowLibrary(me)
			c := openSearch(g, me.ID)
			if c == nil || !c.SearchAsksLibraryZone() {
				t.Fatalf("the restored prompt lost its zone question: %+v", c)
			}
			if err := g.ResolveSearchLibraryZones(c.ID, me.ID, tc.picks(lib, gy), tc.libraryToo); err != nil {
				t.Fatal(err)
			}
			want := 2
			if tc.wantEvents > 0 {
				want = 0
			}
			if searchEvents(g) != tc.wantEvents || libraryKnownTo(me, me.ID) != want {
				t.Errorf("events=%d known=%d, want %d events and %d known", searchEvents(g),
					libraryKnownTo(me, me.ID), tc.wantEvents, want)
			}
		})
	}
}
