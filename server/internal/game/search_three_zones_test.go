package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// search_three_zones_test.go — #5408: "search your graveyard, hand, and/or
// library" (Say Its Name) is AlsoGraveyard with AlsoHand. The zone question
// offers all three (search_other_zone SearchZoneGraveyardAndHand), the
// searcher names any non-empty subset, only the chosen zones' cards are
// offered, and the library is shuffled and the search announced if and
// only if it was chosen (CR 701.23a, 701.24a).

// threeZonesSearch opens the three-zone search for "Wanted" with one match
// in each zone. The library is not yet known to the seat: a test marks it
// known before the take, so a shuffle is observable as un-knowing it. The
// zone question is still open.
func threeZonesSearch(t *testing.T, g *Game) (c *PendingChoice, lib, gy, hand uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	lib = searchLibraryOf(me, "Wanted", "Other")[0]
	gy = gyCard(me, "Wanted")
	hand = uuid.New()
	me.Hand.PushTop(Card{InstanceID: hand, Name: "Wanted", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneBattlefield, Limit: 1, Shuffle: true,
		AlsoGraveyard: true, AlsoHand: true,
	})
	c = openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	return c, lib, gy, hand
}

func TestAThreeZoneSearchAsksWhichOfTheThree(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c, lib, gy, hand := threeZonesSearch(t, g)
	if !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneGraveyardAndHand {
		t.Fatalf("a three-zone search must open on the zone question with the third wire value: %+v", c)
	}
	if searcherKnowsLibraryCard(g, me, lib) {
		t.Fatal("the library's match was made known before the library was chosen")
	}
	for _, picks := range [][]uuid.UUID{nil, {gy}, {hand}, {lib}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a pick %v before the zones are chosen: err = %v, want ErrInvalidParam", picks, err)
		}
	}
	for _, bad := range [][]string{nil, {}, {""}, {SearchZoneGraveyardAndHand}, {"exile"}, {SearchZoneLibrary, "exile"}} {
		if err := g.ChooseSearchZones(c.ID, me.ID, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("zones %q: err = %v, want ErrInvalidParam", bad, err)
		}
	}
	if c := openSearch(g, me.ID); c == nil || !c.SearchZonesPending() {
		t.Fatal("a refused answer must leave the question open")
	}
}

// Every non-empty subset of the three zones: the chosen zones' matches,
// and only those, are offered; each can be taken onto the battlefield; and
// the library is shuffled, and the search announced, if and only if it is
// among them, whatever is taken.
func TestEachThreeZoneSubsetFindsItsMatchAndOnlyTheLibraryShuffles(t *testing.T) {
	subsets := [][]string{
		{SearchZoneLibrary},
		{SearchZoneGraveyard},
		{SearchZoneHand},
		{SearchZoneLibrary, SearchZoneGraveyard},
		{SearchZoneLibrary, SearchZoneHand},
		{SearchZoneGraveyard, SearchZoneHand},
		{SearchZoneLibrary, SearchZoneGraveyard, SearchZoneHand},
	}
	for _, zones := range subsets {
		withLibrary := slices.Contains(zones, SearchZoneLibrary)
		for _, take := range []string{"", SearchZoneLibrary, SearchZoneGraveyard, SearchZoneHand} {
			if take != "" && !slices.Contains(zones, take) {
				continue
			}
			g := newActiveGame(t)
			me := g.Seats[0]
			c, lib, gy, hand := threeZonesSearch(t, g)
			c = answerZones(t, g, c, zones...)
			byZone := map[string]uuid.UUID{SearchZoneLibrary: lib, SearchZoneGraveyard: gy, SearchZoneHand: hand}
			for z, id := range byZone {
				if slices.Contains(c.SearchCards, id) != slices.Contains(zones, z) {
					t.Errorf("zones %v: the %s match offered = %v", zones, z, !slices.Contains(zones, z))
				}
			}
			if slices.Contains(c.SearchGraveyardCards, gy) != slices.Contains(zones, SearchZoneGraveyard) ||
				slices.Contains(c.SearchHandCards, hand) != slices.Contains(zones, SearchZoneHand) {
				t.Errorf("zones %v: zone marks gy=%v hand=%v", zones, c.SearchGraveyardCards, c.SearchHandCards)
			}
			if c.SearchLibrarySkipped() == withLibrary {
				t.Errorf("zones %v: SearchLibrarySkipped = %v", zones, c.SearchLibrarySkipped())
			}
			if searcherKnowsLibraryCard(g, me, lib) != withLibrary {
				t.Errorf("zones %v: library match known = %v", zones, !withLibrary)
			}
			var picks []uuid.UUID
			if take != "" {
				picks = []uuid.UUID{byZone[take]}
			}
			g.WithWriteLock(func() { knowLibrary(me) })
			if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); err != nil {
				t.Fatalf("zones %v take %q: %v", zones, take, err)
			}
			if take != "" && !g.Battlefield.Contains(byZone[take]) {
				t.Errorf("zones %v: the %s match did not enter the battlefield", zones, take)
			}
			wantEvents, wantKnown := 0, len(me.Library.Cards)
			if withLibrary {
				wantEvents, wantKnown = 1, 0
			}
			if searchEvents(g) != wantEvents || libraryKnownTo(me, me.ID) != wantKnown {
				t.Errorf("zones %v take %q: events=%d known=%d, want %d/%d", zones, take,
					searchEvents(g), libraryKnownTo(me, me.ID), wantEvents, wantKnown)
			}
		}
	}
}

// Nothing matches anywhere: the question and the pick are still asked
// (#4140), finding nothing is legal, and the library is shuffled only if it
// was searched.
func TestAnEmptyThreeZoneSearchFindsNothing(t *testing.T) {
	for _, zones := range [][]string{
		{SearchZoneGraveyard, SearchZoneHand},
		{SearchZoneLibrary, SearchZoneGraveyard, SearchZoneHand},
	} {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me, "Other", "Other")
		runSearch(t, g, SearchLibrarySpec{
			Player: me.ID, Pred: named("Wanted"), Dest: ZoneBattlefield, Limit: 1, Shuffle: true,
			AlsoGraveyard: true, AlsoHand: true,
		})
		c := openSearch(g, me.ID)
		if c == nil || !c.SearchZonesPending() {
			t.Fatalf("an empty three-zone search must still ask: %+v", c)
		}
		c = answerZones(t, g, c, zones...)
		if len(c.SearchCards) != 0 || c.SearchMin != 0 {
			t.Fatalf("zones %v: offered %v, floor %d", zones, c.SearchCards, c.SearchMin)
		}
		g.WithWriteLock(func() { knowLibrary(me) })
		if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
			t.Fatal(err)
		}
		want := 0
		if slices.Contains(zones, SearchZoneLibrary) {
			want = 1
		}
		if searchEvents(g) != want {
			t.Errorf("zones %v: events=%d, want %d", zones, searchEvents(g), want)
		}
	}
}

// A restore keeps the three-zone question, and an answered one keeps its
// trimmed zones.
func TestARestoreKeepsTheThreeZoneQuestion(t *testing.T) {
	g := newActiveGame(t)
	threeZonesSearch(t, g)
	g = strictSearchRestore(t, g)
	me := g.Seats[0]
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneGraveyardAndHand {
		t.Fatalf("the restored prompt lost its three-zone question: %+v", c)
	}

	g = newActiveGame(t)
	c, lib, gy, hand := threeZonesSearch(t, g)
	answerZones(t, g, c, SearchZoneHand)
	g = strictSearchRestore(t, g)
	me = g.Seats[0]
	c = openSearch(g, me.ID)
	if c == nil || c.SearchZonesPending() || !c.SearchLibrarySkipped() {
		t.Fatalf("the restored prompt must stay answered, library skipped: %+v", c)
	}
	if !slices.Equal(c.SearchCards, []uuid.UUID{hand}) || slices.Contains(c.SearchCards, gy) || slices.Contains(c.SearchCards, lib) {
		t.Errorf("restored candidates %v, want only the hand match", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{hand}); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 0 {
		t.Errorf("a hand-only take on the restored table announced %d searches", searchEvents(g))
	}
}
