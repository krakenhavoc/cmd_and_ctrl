package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// search_zones_test.go — #4516, #4519: an and/or search ("your library
// and/or graveyard", "your hand and/or library") asks which zones the
// searcher looks in BEFORE it shows anything. The printed text shuffles
// only "if you search your library this way" (CR 701.23), and looking
// through the library is searching it (CR 701.23a): the library's
// candidates are shown, and known to the searcher, only if the library is
// chosen, and choosing it is what shuffles.

// zonesSearch opens an "and/or graveyard" search for "Wanted" with one
// match in each zone and the library known to the seat (so a shuffle is
// observable as un-knowing it). The zone question is still open. It
// returns the prompt, the library match and the graveyard match.
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

// handZonesSearch is zonesSearch for "search your hand and/or library":
// one match in the library, one in the hand.
func handZonesSearch(t *testing.T, g *Game) (c *PendingChoice, lib, hand uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	lib = searchLibraryOf(me, "Wanted", "Other")[0]
	knowLibrary(me)
	hand = uuid.New()
	me.Hand.PushTop(Card{InstanceID: hand, Name: "Wanted", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneBattlefield, Limit: 1, Shuffle: true, AlsoHand: true,
	})
	c = openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	return c, lib, hand
}

// answerZones answers the open zone question and returns the prompt as
// it stands after.
func answerZones(t *testing.T, g *Game, c *PendingChoice, zones ...string) *PendingChoice {
	t.Helper()
	who := g.Seats[0].ID
	if err := g.ChooseSearchZones(c.ID, who, zones); err != nil {
		t.Fatalf("ChooseSearchZones(%v): %v", zones, err)
	}
	c = openSearch(g, who)
	if c == nil {
		t.Fatal("the zone answer closed the prompt")
	}
	if c.SearchZonesPending() {
		t.Fatal("the prompt is still asking the zones after the answer")
	}
	return c
}

// searcherKnowsLibraryCard reports whether the seat is a knower of id,
// a card in its library.
func searcherKnowsLibraryCard(g *Game, p *Player, id uuid.UUID) bool {
	known := false
	g.ReadSnapshot(func() {
		for _, c := range p.Library.Cards {
			if c.InstanceID == id {
				known = c.KnownBy[p.ID]
			}
		}
	})
	return known
}

// The question comes first: before it is answered the prompt takes no
// pick, grants no knowledge of the library's candidates, and says what it
// is asking.
func TestAnAndOrSearchAsksTheZonesFirst(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Wanted", "Other")
	lib := me.Library.Cards[0].InstanceID
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneGraveyard {
		t.Fatalf("an \"and/or graveyard\" search must open on the zone question: %+v", c)
	}
	if searcherKnowsLibraryCard(g, me, lib) {
		t.Fatal("the library's match was made known before the searcher chose to search the library")
	}
	for _, picks := range [][]uuid.UUID{nil, {gy}, {lib}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a pick %v before the zones are chosen: err = %v, want ErrInvalidParam", picks, err)
		}
	}
	if g.SearchPickLegalLocked(c, []uuid.UUID{gy}) {
		t.Error("the enumerator's acceptance check must refuse a pick before the zones are chosen")
	}
	if openSearch(g, me.ID) == nil {
		t.Error("a refused answer must leave the prompt open")
	}
}

// Bad answers to the question are refused and leave it open.
func TestAnAndOrSearchRefusesABadZoneAnswer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c, _, _ := zonesSearch(t, g, SearchLibrarySpec{})
	for name, zones := range map[string][]string{
		"empty":                    {},
		"nil":                      nil,
		"a zone this prompt lacks": {SearchZoneHand},
		"an unknown zone":          {"exile"},
		"library and a bad one":    {SearchZoneLibrary, "stack"},
	} {
		if err := g.ChooseSearchZones(c.ID, me.ID, zones); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err = %v, want ErrInvalidParam", name, err)
		}
	}
	if err := g.ChooseSearchZones(c.ID, g.Seats[1].ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat answering: err = %v, want ErrNotTheChooser", err)
	}
	if err := g.ChooseSearchZones(uuid.New(), me.ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrPendingChoiceNotFound) {
		t.Errorf("an unknown prompt: err = %v, want ErrPendingChoiceNotFound", err)
	}
	if c := openSearch(g, me.ID); c == nil || !c.SearchZonesPending() {
		t.Fatal("a refused answer must leave the question open")
	}
	// Once answered it cannot be asked again.
	answerZones(t, g, c, SearchZoneGraveyard)
	if err := g.ChooseSearchZones(c.ID, me.ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("answering twice: err = %v, want ErrInvalidParam", err)
	}
}

// A search with no zone to choose never asks.
func TestOnlyAndOrSearchesAskWhichZones(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	knowLibrary(me)
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true})
	c := openSearch(g, me.ID)
	if c.SearchZonesPending() || c.SearchOtherZone() != "" {
		t.Error("a library-only search has no zone to ask about")
	}
	if err := g.ChooseSearchZones(c.ID, me.ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a zone answer to a library-only search: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
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
	if c.SearchZonesPending() {
		t.Error("\"graveyard, hand, and library\" names every zone, so there is nothing to ask")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: the all-zones search shuffles", searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// Graveyard only: the library's candidates are never offered and never
// made known, a library card is not a legal pick, and there is no shuffle
// and no search event, whether the searcher takes the graveyard card or
// nothing.
func TestGraveyardOnlyNeverOffersTheLibraryAndDoesNotShuffle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		takes bool
	}{{"takes the graveyard card", true}, {"finds nothing", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			c, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
			// The library match is unknown to the seat until it is chosen.
			g.WithWriteLock(func() {
				for i := range me.Library.Cards {
					delete(me.Library.Cards[i].KnownBy, me.ID)
				}
			})
			c = answerZones(t, g, c, SearchZoneGraveyard)
			if !slices.Equal(c.SearchCards, []uuid.UUID{gy}) || !slices.Equal(c.SearchGraveyardCards, []uuid.UUID{gy}) {
				t.Fatalf("offered %v (graveyard %v), want only the graveyard card", c.SearchCards, c.SearchGraveyardCards)
			}
			if searcherKnowsLibraryCard(g, me, lib) {
				t.Error("the library's match was made known to a searcher who did not search the library")
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lib}); !errors.Is(err, ErrInvalidParam) {
				t.Errorf("taking a library card from a graveyard-only search: err = %v, want ErrInvalidParam", err)
			}
			// Mark the library again so a shuffle is observable.
			g.WithWriteLock(func() { knowLibrary(me) })
			var picks []uuid.UUID
			if tc.takes {
				picks = []uuid.UUID{gy}
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); err != nil {
				t.Fatal(err)
			}
			if tc.takes && !me.Hand.Contains(gy) {
				t.Fatal("the graveyard card did not reach the hand")
			}
			if searchEvents(g) != 0 || libraryKnownTo(me, me.ID) != 2 {
				t.Errorf("events=%d known=%d: a graveyard-only search must not shuffle or announce",
					searchEvents(g), libraryKnownTo(me, me.ID))
			}
		})
	}
}

// The library chosen: its candidates are offered beside the graveyard's and
// become known to the searcher, and the library is shuffled whatever is
// taken: the library card, the graveyard card, or nothing.
func TestChoosingTheLibraryOffersItAndShuffles(t *testing.T) {
	for _, tc := range []struct {
		name  string
		zones []string
		picks func(lib, gy uuid.UUID) []uuid.UUID
	}{
		{"library only, takes the library card", []string{SearchZoneLibrary}, func(lib, _ uuid.UUID) []uuid.UUID { return []uuid.UUID{lib} }},
		{"library only, finds nothing", []string{SearchZoneLibrary}, func(_, _ uuid.UUID) []uuid.UUID { return nil }},
		{"both, takes the graveyard card", []string{SearchZoneLibrary, SearchZoneGraveyard}, func(_, gy uuid.UUID) []uuid.UUID { return []uuid.UUID{gy} }},
		{"both, takes the library card", []string{SearchZoneGraveyard, SearchZoneLibrary}, func(lib, _ uuid.UUID) []uuid.UUID { return []uuid.UUID{lib} }},
		{"both, finds nothing", []string{SearchZoneLibrary, SearchZoneGraveyard}, func(_, _ uuid.UUID) []uuid.UUID { return nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			c, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
			g.WithWriteLock(func() {
				for i := range me.Library.Cards {
					delete(me.Library.Cards[i].KnownBy, me.ID)
				}
			})
			c = answerZones(t, g, c, tc.zones...)
			if !slices.Contains(c.SearchCards, lib) {
				t.Fatalf("the chosen library's match was not offered: %v", c.SearchCards)
			}
			if slices.Contains(c.SearchCards, gy) != slices.Contains(tc.zones, SearchZoneGraveyard) {
				t.Errorf("the graveyard match is offered = %v with zones %v", slices.Contains(c.SearchCards, gy), tc.zones)
			}
			if !searcherKnowsLibraryCard(g, me, lib) {
				t.Error("the searcher does not know the library card they chose to search for")
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, tc.picks(lib, gy)); err != nil {
				t.Fatal(err)
			}
			if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
				t.Errorf("events=%d known=%d: a library search shuffles and is announced",
					searchEvents(g), libraryKnownTo(me, me.ID))
			}
		})
	}
}

// A graveyard card onto the battlefield finishes through the entry
// pipeline's tail; the zones answer rides it.
func TestGraveyardPickToTheBattlefieldKeepsTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		zones      []string
		wantEvents int
	}{
		{[]string{SearchZoneGraveyard}, 0},
		{[]string{SearchZoneLibrary, SearchZoneGraveyard}, 1},
	} {
		g := newActiveGame(t)
		me := g.Seats[0]
		c, _, gy := zonesSearch(t, g, SearchLibrarySpec{Dest: ZoneBattlefield})
		c = answerZones(t, g, c, tc.zones...)
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gy}); err != nil {
			t.Fatal(err)
		}
		if !g.Battlefield.Contains(gy) {
			t.Fatal("the graveyard card did not enter the battlefield")
		}
		if searchEvents(g) != tc.wantEvents {
			t.Errorf("zones %v: events=%d, want %d", tc.zones, searchEvents(g), tc.wantEvents)
		}
	}
}

// "Search your hand and/or library" (Last Light of Durin's Day) asks the
// same question, with the hand as the other zone.
func TestAHandAndOrLibrarySearchAsksTheZonesToo(t *testing.T) {
	t.Run("hand only does not shuffle and never offers the library", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		c, lib, hand := handZonesSearch(t, g)
		if !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneHand {
			t.Fatalf("a \"hand and/or library\" search must ask: %+v", c)
		}
		if err := g.ChooseSearchZones(c.ID, me.ID, []string{SearchZoneGraveyard}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a graveyard answer to a hand search: err = %v, want ErrInvalidParam", err)
		}
		g.WithWriteLock(func() {
			for i := range me.Library.Cards {
				delete(me.Library.Cards[i].KnownBy, me.ID)
			}
		})
		c = answerZones(t, g, c, SearchZoneHand)
		if !slices.Equal(c.SearchCards, []uuid.UUID{hand}) || !slices.Equal(c.SearchHandCards, []uuid.UUID{hand}) {
			t.Fatalf("offered %v (hand %v), want only the hand card", c.SearchCards, c.SearchHandCards)
		}
		if searcherKnowsLibraryCard(g, me, lib) {
			t.Error("the library's match was made known to a searcher who searched only their hand")
		}
		g.WithWriteLock(func() { knowLibrary(me) })
		if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
			t.Fatal(err)
		}
		if searchEvents(g) != 0 || libraryKnownTo(me, me.ID) != 2 {
			t.Errorf("events=%d known=%d: a hand-only search that finds nothing must not shuffle",
				searchEvents(g), libraryKnownTo(me, me.ID))
		}
	})
	t.Run("hand only, taking the hand card, does not shuffle", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		c, _, hand := handZonesSearch(t, g)
		c = answerZones(t, g, c, SearchZoneHand)
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{hand}); err != nil {
			t.Fatal(err)
		}
		if !g.Battlefield.Contains(hand) {
			t.Fatal("the hand card did not enter the battlefield")
		}
		if searchEvents(g) != 0 || libraryKnownTo(me, me.ID) != 2 {
			t.Errorf("events=%d known=%d: a hand-only take must not shuffle", searchEvents(g), libraryKnownTo(me, me.ID))
		}
	})
	t.Run("the library chosen shuffles, taking nothing included", func(t *testing.T) {
		for _, zones := range [][]string{{SearchZoneLibrary}, {SearchZoneLibrary, SearchZoneHand}} {
			g := newActiveGame(t)
			me := g.Seats[0]
			c, lib, _ := handZonesSearch(t, g)
			c = answerZones(t, g, c, zones...)
			if !slices.Contains(c.SearchCards, lib) {
				t.Fatalf("zones %v: the library match was not offered", zones)
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
				t.Fatal(err)
			}
			if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
				t.Errorf("zones %v: events=%d known=%d, want a shuffle", zones, searchEvents(g), libraryKnownTo(me, me.ID))
			}
		}
	})
}

// Nothing in either zone matches. The question is still asked, and so is
// the pick that follows it: another seat sees the same two prompts whether
// or not the library held a match (#4140).
func TestAnEmptyAndOrSearchStillAsksBothQuestions(t *testing.T) {
	for _, tc := range []struct {
		zones      []string
		wantEvents int
		wantKnown  int
	}{
		{[]string{SearchZoneGraveyard}, 0, 2},
		{[]string{SearchZoneLibrary}, 1, 0},
		{[]string{SearchZoneLibrary, SearchZoneGraveyard}, 1, 0},
	} {
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
		if c == nil || !c.SearchZonesPending() {
			t.Fatalf("an empty search still asks the zones: %+v", c)
		}
		c = answerZones(t, g, c, tc.zones...)
		if len(c.SearchCards) != 0 {
			t.Fatalf("offered %v, want nothing", c.SearchCards)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
			t.Fatal(err)
		}
		if searchEvents(g) != tc.wantEvents || libraryKnownTo(me, me.ID) != tc.wantKnown {
			t.Errorf("zones %v: events=%d known=%d, want %d/%d", tc.zones, searchEvents(g),
				libraryKnownTo(me, me.ID), tc.wantEvents, tc.wantKnown)
		}
		if !ran {
			t.Errorf("zones %v: the search's \"then\" did not run with nothing found", tc.zones)
		}
	}
}

// An unconditional "a card" and/or search must find min(limit, cards in
// the chosen zones): the floor follows the zones chosen, and the engine
// never takes the match unasked, since the zones are the player's to pick.
func TestTheFloorFollowsTheChosenZones(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true})
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchZonesPending() {
		t.Fatalf("an unconditional and/or search with a single card per zone must still ask: %+v", c)
	}
	c = answerZones(t, g, c, SearchZoneGraveyard)
	if c.SearchMin != 1 || !slices.Equal(c.SearchCards, []uuid.UUID{gy}) {
		t.Fatalf("graveyard only: min=%d cards=%v, want 1 and the graveyard card", c.SearchMin, c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("finding nothing when the chosen zone holds a card: err = %v, want ErrInvalidParam", err)
	}
}

// A restore at each step keeps the question, and the answers given on the
// restored table decide the shuffle exactly as on the live one.
func TestARestoreAtEachStepOfTheZoneQuestion(t *testing.T) {
	t.Run("while the question is open", func(t *testing.T) {
		for _, tc := range []struct {
			zones      []string
			wantEvents int
			wantKnown  int
			wantOffers bool
		}{
			{[]string{SearchZoneGraveyard}, 0, 2, false},
			{[]string{SearchZoneLibrary}, 1, 0, true},
			{[]string{SearchZoneLibrary, SearchZoneGraveyard}, 1, 0, true},
		} {
			g := newActiveGame(t)
			_, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
			g = strictSearchRestore(t, g)
			me := g.Seats[0]
			knowLibrary(me)
			c := openSearch(g, me.ID)
			if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneGraveyard {
				t.Fatalf("the restored prompt lost its zone question: %+v", c)
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gy}); !errors.Is(err, ErrInvalidParam) {
				t.Errorf("a pick on the restored, unanswered prompt: err = %v, want ErrInvalidParam", err)
			}
			c = answerZones(t, g, c, tc.zones...)
			if slices.Contains(c.SearchCards, lib) != tc.wantOffers {
				t.Errorf("zones %v: library match offered = %v, want %v", tc.zones, !tc.wantOffers, tc.wantOffers)
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
				t.Fatal(err)
			}
			if searchEvents(g) != tc.wantEvents || libraryKnownTo(me, me.ID) != tc.wantKnown {
				t.Errorf("zones %v: events=%d known=%d, want %d/%d", tc.zones, searchEvents(g),
					libraryKnownTo(me, me.ID), tc.wantEvents, tc.wantKnown)
			}
		}
	})
	t.Run("after the zones are chosen", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			zones      []string
			picks      func(lib, gy uuid.UUID) []uuid.UUID
			wantEvents int
		}{
			{"graveyard only, nothing", []string{SearchZoneGraveyard}, func(_, _ uuid.UUID) []uuid.UUID { return nil }, 0},
			{"graveyard only, the graveyard card", []string{SearchZoneGraveyard}, func(_, gy uuid.UUID) []uuid.UUID { return []uuid.UUID{gy} }, 0},
			{"library, nothing", []string{SearchZoneLibrary}, func(_, _ uuid.UUID) []uuid.UUID { return nil }, 1},
			{"both, the graveyard card", []string{SearchZoneLibrary, SearchZoneGraveyard}, func(_, gy uuid.UUID) []uuid.UUID { return []uuid.UUID{gy} }, 1},
			{"library, the library card", []string{SearchZoneLibrary}, func(lib, _ uuid.UUID) []uuid.UUID { return []uuid.UUID{lib} }, 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := newActiveGame(t)
				c, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
				answerZones(t, g, c, tc.zones...)
				g = strictSearchRestore(t, g)
				me := g.Seats[0]
				knowLibrary(me)
				c = openSearch(g, me.ID)
				if c == nil || c.SearchZonesPending() {
					t.Fatalf("the restored prompt must stay answered: %+v", c)
				}
				if slices.Contains(c.SearchCards, lib) != slices.Contains(tc.zones, SearchZoneLibrary) {
					t.Errorf("the restored candidates %v do not follow the zones %v", c.SearchCards, tc.zones)
				}
				if err := g.ChooseSearchZones(c.ID, me.ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrInvalidParam) {
					t.Errorf("re-asking an answered question on the restored table: err = %v, want ErrInvalidParam", err)
				}
				if err := g.ResolveSearchLibrary(c.ID, me.ID, tc.picks(lib, gy)); err != nil {
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
	})
	t.Run("a hand search", func(t *testing.T) {
		g := newActiveGame(t)
		handZonesSearch(t, g)
		g = strictSearchRestore(t, g)
		me := g.Seats[0]
		c := openSearch(g, me.ID)
		if c == nil || !c.SearchZonesPending() || c.SearchOtherZone() != SearchZoneHand {
			t.Fatalf("the restored hand prompt lost its zone question: %+v", c)
		}
	})
}

// A "you may search" that is also an and/or search asks "do you search?"
// first and the zones second; neither shows a library card.
func TestAnOptionalAndOrSearchAsksWhetherThenWhichZones(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lib := searchLibraryOf(me, "Wanted", "Other")[0]
	gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true, Optional: true,
	})
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchIsUndecided() || !c.SearchZonesPending() {
		t.Fatalf("expected the whether-question with the zone question behind it: %+v", c)
	}
	if err := g.ChooseSearchZones(c.ID, me.ID, []string{SearchZoneLibrary}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the zones before whether: err = %v, want ErrInvalidParam", err)
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatal(err)
	}
	c = openSearch(g, me.ID)
	if c.SearchIsUndecided() || !c.SearchZonesPending() {
		t.Fatalf("after \"search\" the zone question must still be open: %+v", c)
	}
	if searcherKnowsLibraryCard(g, me, lib) {
		t.Error("saying yes to the search showed the library before the zones were chosen")
	}
	answerZones(t, g, c, SearchZoneLibrary)
	if !searcherKnowsLibraryCard(g, me, lib) {
		t.Error("the library's match must be known once the library is chosen")
	}
}

// An unconditional and/or search with exactly as many matches as it must
// find would be taken without a prompt for any other search (CR 701.23d).
// Not this one: which zones it searches is the player's decision, and the
// library's shuffle follows it.
func TestAnUnconditionalAndOrSearchIsNeverTakenUnasked(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lib := searchLibraryOf(me, "Only")[0]
	knowLibrary(me)
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true})
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchZonesPending() {
		t.Fatalf("the single match was taken without asking the zones: %+v", c)
	}
	if me.Hand.Contains(lib) {
		t.Fatal("the library card moved before the zones were chosen")
	}
	c = answerZones(t, g, c, SearchZoneLibrary)
	if c.SearchMin != 1 || !slices.Equal(c.SearchCards, []uuid.UUID{lib}) {
		t.Fatalf("min=%d cards=%v, want the one library card as a must-find", c.SearchMin, c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lib}); err != nil {
		t.Fatal(err)
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: a library search shuffles", searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// The zone step is written with the prompt, under a key of its own, so a
// binary that does not know it refuses the file instead of showing the
// library before the zones were chosen.
func TestTheZoneStepIsWrittenUnderItsOwnKey(t *testing.T) {
	g := newActiveGame(t)
	zonesSearch(t, g, SearchLibrarySpec{})
	snap := throughJSON(t, g.CaptureSnapshot())
	var found bool
	for _, c := range snap.PendingChoices {
		if c.Kind != PendingChoiceSearchLibrary {
			continue
		}
		found = true
		if c.SearchPrompt == nil || c.SearchPrompt.Zones != "asking" || c.PickThen != SearchPromptZonesKey {
			t.Fatalf("zones=%+v pickThen=%q, want asking under %q", c.SearchPrompt, c.PickThen, SearchPromptZonesKey)
		}
		if KnownRevealedPickThen(SearchPromptZonesKey) || SearchPromptZonesKey == SearchPromptKey || SearchPromptZonesKey == SearchPromptWordsKey {
			t.Error("the zones key must be its own on-disk identity")
		}
	}
	if !found {
		t.Fatal("the search prompt was not written")
	}
}

// A prompt written before the zone question existed (no zones field, the
// plain key) restores as a search that does not ask, and keeps the reading
// #4516 gave it: a take wholly out of the graveyard does not shuffle.
func TestAPromptWrittenBeforeTheZoneQuestionStillRestores(t *testing.T) {
	g := newActiveGame(t)
	_, lib, gy := zonesSearch(t, g, SearchLibrarySpec{})
	snap := throughJSON(t, g.CaptureSnapshot())
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].Kind == PendingChoiceSearchLibrary {
			snap.PendingChoices[i].SearchPrompt.Zones = ""
			snap.PendingChoices[i].PickThen = SearchPromptKey
		}
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	me := restored.Seats[0]
	knowLibrary(me)
	c := openSearch(restored, me.ID)
	if c == nil || c.SearchZonesPending() {
		t.Fatalf("an old prompt must not start asking: %+v", c)
	}
	if !slices.Contains(c.SearchCards, lib) || !slices.Contains(c.SearchCards, gy) {
		t.Fatalf("an old prompt lost candidates: %v", c.SearchCards)
	}
	if err := restored.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gy}); err != nil {
		t.Fatal(err)
	}
	if searchEvents(restored) != 0 || libraryKnownTo(me, me.ID) != 2 {
		t.Errorf("events=%d known=%d: the old reading of a graveyard take is no shuffle",
			searchEvents(restored), libraryKnownTo(me, me.ID))
	}
}
