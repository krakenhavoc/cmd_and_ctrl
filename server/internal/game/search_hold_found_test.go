package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_hold_found_test.go — #6156: "search your library and graveyard
// ... and reveal them" whose found cards wait where they were found
// (SearchLibrarySpec.HoldFound) while the rest of the effect splits them,
// and "library AND graveyard" naming both zones (BothZones).
//
// CR 701.23a: to search a zone, look at all cards in it and find cards
// that match. CR 701.20a: a revealed card stays revealed for as long as
// the parts of the effect it is relevant to take.

// HoldFound leaves the found cards in the library and the graveyard and
// names them, in the order taken, to the rest of the effect.
func TestHoldFoundLeavesTheCardsWhereTheyWereAndNamesThem(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lib := searchLibraryOf(me, "Wanted A", "Other", "Wanted B")
	gy := gyCard(me, "Wanted C")
	var then []uuid.UUID
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: func(c Card) bool { return c.Name != "Other" }, Dest: ZoneLibrary,
		HoldFound: true, AlsoGraveyard: true, BothZones: true, Limit: 4, Reveal: true,
		ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { then = found; return nil }),
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if c.SearchZonesPending() {
		t.Fatal("\"library and graveyard\" names both zones: no zone question")
	}
	if len(c.SearchCards) != 3 {
		t.Fatalf("candidates = %v, want both library matches and the graveyard match", c.SearchCards)
	}
	pick := []uuid.UUID{lib[0], gy, lib[2]}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, pick); err != nil {
		t.Fatal(err)
	}
	if len(then) != 3 || then[0] != lib[0] || then[1] != gy || then[2] != lib[2] {
		t.Fatalf("the rest of the effect was told %v, want %v", then, pick)
	}
	if !me.Library.Contains(lib[0]) || !me.Library.Contains(lib[2]) || !me.Graveyard.Contains(gy) {
		t.Error("a found card moved: HoldFound leaves each where it was found")
	}
	if !holdFoundKnownTo(me, lib[0], g.Seats[1].ID) {
		t.Error("a found library card was not revealed to the table")
	}
	if holdFoundKnownTo(me, lib[1], g.Seats[1].ID) {
		t.Error("an unfound library card was revealed")
	}
	if searchEvents(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: the library was searched", searchEvents(g))
	}
}

// BothZones: a take wholly out of the graveyard, or nothing at all, is
// still a search of the library, because the card names it.
func TestLibraryAndGraveyardIsALibrarySearchWhateverIsTaken(t *testing.T) {
	for _, takeGraveyard := range []bool{true, false} {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me, "Wanted")
		gy := gyCard(me, "Wanted")
		runSearch(t, g, SearchLibrarySpec{
			Player: me.ID, Pred: named("Wanted"), Dest: ZoneLibrary, HoldFound: true,
			AlsoGraveyard: true, BothZones: true, Limit: 2, Reveal: true,
		})
		c := openSearch(g, me.ID)
		if c == nil {
			t.Fatal("no prompt")
		}
		var pick []uuid.UUID
		if takeGraveyard {
			pick = []uuid.UUID{gy}
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, pick); err != nil {
			t.Fatal(err)
		}
		if searchEvents(g) != 1 {
			t.Errorf("take graveyard=%v: EventSearchLibrary fired %d times, want 1", takeGraveyard, searchEvents(g))
		}
	}
}

// Without HoldFound a "leave it where it is" search still names nothing to
// the rest of the effect: the field changes only the cards that opt in.
func TestLeaveInPlaceSearchWithoutHoldFoundNamesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lib := searchLibraryOf(me, "Wanted")
	called, got := false, []uuid.UUID(nil)
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneLibrary, Limit: 1, Reveal: true,
		ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { called, got = true, found; return nil }),
	})
	c := openSearch(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, lib); err != nil {
		t.Fatal(err)
	}
	if !called || len(got) != 0 {
		t.Errorf("called=%v found=%v, want called with nothing", called, got)
	}
}

// holdFoundKnownTo reports whether viewer knows the library card id.
func holdFoundKnownTo(p *Player, id, viewer uuid.UUID) bool {
	for i := range p.Library.Cards {
		if p.Library.Cards[i].InstanceID == id {
			return p.Library.Cards[i].IsKnownTo(viewer)
		}
	}
	return false
}
