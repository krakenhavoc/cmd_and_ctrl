package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// search_graveyard_test.go — #3217: "search your library and/or
// graveyard" (SearchLibrarySpec.AlsoGraveyard), the rule beside the
// code. Tower Winder, Finale of Devastation and Ajani's Aid exercise it
// through the catalog.
//
// "And/or" lets the searcher choose their zones, so a graveyard match
// does not remove the right to find nothing; only an unconditional
// search has a floor.

func TestFloorZonesTable(t *testing.T) {
	named := func(c Card) bool { return c.Name == "X" }
	cases := []struct {
		name   string
		spec   SearchLibrarySpec
		lib, g int
		want   int
	}{
		{"stated quality, library only", SearchLibrarySpec{Pred: named, Limit: 1}, 3, 0, 0},
		{"stated quality, graveyard match", SearchLibrarySpec{Pred: named, Limit: 1}, 0, 1, 0},
		{"stated quality, both zones", SearchLibrarySpec{Pred: named, Limit: 1}, 5, 1, 0},
		{"stated quality, two wanted, one in the graveyard", SearchLibrarySpec{Pred: named, Limit: 2}, 5, 1, 0},
		{"stated quality, up to five", SearchLibrarySpec{Pred: named, Limit: 5, UpTo: true}, 5, 3, 0},
		{"unconditional counts both zones", SearchLibrarySpec{Limit: 3}, 1, 1, 2},
		{"unconditional, empty library, graveyard card", SearchLibrarySpec{Limit: 1}, 0, 1, 1},
		{"unconditional, up to three", SearchLibrarySpec{Limit: 3, UpTo: true}, 1, 1, 0},
		{"you may search", SearchLibrarySpec{Pred: named, Limit: 1, Optional: true}, 1, 1, 0},
	}
	for _, c := range cases {
		if got := c.spec.floorZones(c.lib, c.g); got != c.want {
			t.Errorf("%s: floorZones(%d,%d) = %d, want %d", c.name, c.lib, c.g, got, c.want)
		}
	}
}

func gyCard(p *Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", Owner: p.ID, Controller: p.ID})
	return id
}

func knowLibrary(p *Player) {
	for i := range p.Library.Cards {
		p.Library.Cards[i].AddKnower(p.ID)
	}
}

func named(n string) func(Card) bool { return func(c Card) bool { return c.Name == n } }

// A graveyard match, nothing in the library: the searcher takes it, no
// search event, no shuffle, Then runs with the card.
func TestGraveyardOnlyMatchIsFoundWithoutSearchingTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	knowLibrary(me)
	id := gyCard(me, "Wanted")
	var then []uuid.UUID
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true,
		AlsoGraveyard: true, ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { then = found; return nil }),
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	c = answerZones(t, g, c, SearchZoneGraveyard)
	if c.SearchMin != 0 {
		t.Fatalf("a stated-quality search asks and may find nothing: %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{id}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the graveyard card did not reach the hand")
	}
	if len(then) != 1 || then[0] != id {
		t.Errorf("Then saw %v, want the graveyard card", then)
	}
	if searchEvents(g) != 0 {
		t.Errorf("EventSearchLibrary fired %d times for a graveyard-only take", searchEvents(g))
	}
	if libraryKnownTo(me, me.ID) != 1 {
		t.Error("the library was shuffled although it was not searched")
	}
}

// A graveyard take onto the battlefield goes through the entry pipeline
// with OldZone graveyard, under the searcher's control.
func TestGraveyardTakeToTheBattlefieldEnters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	id := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneBattlefield, Limit: 1, Shuffle: true,
		AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	c = answerZones(t, g, c, SearchZoneGraveyard)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{id}); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the graveyard card did not enter the battlefield")
	}
	entered := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventETB && ev.CardID == id {
				entered = true
			}
		}
	})
	if !entered {
		t.Error("no ETB event for the card that entered from the graveyard")
	}
}

// Both zones: both offered and marked; the floor stays zero.
func TestBothZonesAreOfferedMarkedByZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Wanted", "Other")
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	c = answerZones(t, g, c, SearchZoneLibrary, SearchZoneGraveyard)
	if len(c.SearchCards) != 2 || c.SearchCards[0] != ids[0] || c.SearchCards[1] != gy {
		t.Errorf("candidates = %v, want library match then graveyard match", c.SearchCards)
	}
	if len(c.SearchGraveyardCards) != 1 || c.SearchGraveyardCards[0] != gy {
		t.Errorf("graveyard candidates = %v, want [%s]", c.SearchGraveyardCards, gy)
	}
	if c.SearchMin != 0 {
		t.Errorf("SearchMin = %d, want 0: \"and/or\" lets the searcher look only in the library", c.SearchMin)
	}
	g.ReadSnapshot(func() {
		if got := g.SearchFloorLocked(c); got != 0 {
			t.Errorf("SearchFloorLocked = %d, want 0", got)
		}
	})
}

// Finding nothing with a graveyard match present is legal. The searcher
// chose the graveyard alone (#4516, #4519), so the event does not fire
// and the library is not shuffled.
func TestFindingNothingWithAGraveyardMatchIsAllowed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Wanted")
	knowLibrary(me)
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	c = answerZones(t, g, c, SearchZoneGraveyard)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing = %v, want accepted", err)
	}
	if !me.Graveyard.Contains(gy) || me.Hand.Contains(gy) {
		t.Error("nothing should have been taken")
	}
	if searchEvents(g) != 0 || libraryKnownTo(me, me.ID) != 1 {
		t.Errorf("events=%d known=%d: finding nothing in the graveyard alone is not a library search",
			searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// An unconditional "a card" search is the one that keeps a floor, and it
// counts BOTH zones. The floor is recomputed from the live zones, so it
// falls when the candidates leave rather than wedging the table.
func TestUnconditionalAlsoGraveyardFloorCountsBothZonesAndFalls(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lib := searchLibraryOf(me, "A")
	gy := gyCard(me, "B")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	c = answerZones(t, g, c, SearchZoneLibrary, SearchZoneGraveyard)
	if len(c.SearchCards) != 2 || c.SearchMin != 1 {
		t.Fatalf("candidates/min = %d/%d, want 2/1", len(c.SearchCards), c.SearchMin)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("an unconditional search must find: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, gy); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneLibrary, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, lib[0]); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("both candidates are gone, so finding nothing must be accepted: %v", err)
	}
}

// Taking one of each from a two-card search: any library card means the
// library was searched, so it shuffles.
func TestMixedTakeShuffles(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Wanted")
	knowLibrary(me)
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 2, Shuffle: true, AlsoGraveyard: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		// Two wanted of two available, but a stated quality has floor 0,
		// not the match count, so this asks.
		t.Fatal("expected a prompt")
	}
	c = answerZones(t, g, c, SearchZoneLibrary, SearchZoneGraveyard)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], gy}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(gy) {
		t.Fatal("both cards should be in hand")
	}
	if searchEvents(g) != 1 || libraryKnownTo(me, me.ID) != 0 {
		t.Errorf("events=%d known=%d: any library take is a library search", searchEvents(g), libraryKnownTo(me, me.ID))
	}
}

// Without the flag nothing changes: the graveyard is not read.
func TestWithoutAlsoGraveyardTheGraveyardIsIgnored(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Other")
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true})
	// #4140: nothing in the library matches, so the prompt is empty.
	c := openSearch(g, me.ID)
	if c == nil || len(c.SearchCards) != 0 {
		t.Fatalf("a library-only search must not see the graveyard: %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Contains(gy) {
		t.Error("a library-only search must not see the graveyard")
	}
	if searchEvents(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", searchEvents(g))
	}
}

// The prompt survives Clone (undo) with its graveyard half.
func TestCloneCarriesTheGraveyardCandidates(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Wanted")
	gy := gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	cl := g.Clone()
	c := openSearch(cl, me.ID)
	if c == nil || len(c.SearchGraveyardCards) != 1 || c.SearchGraveyardCards[0] != gy {
		t.Fatalf("clone lost the graveyard half: %+v", c)
	}
	if !c.SearchZonesPending() {
		t.Error("the clone lost the open zone question")
	}
	orig := openSearch(g, me.ID)
	c.SearchGraveyardCards[0] = uuid.New()
	if orig.SearchGraveyardCards[0] != gy {
		t.Error("the clone shares its backing array with the live prompt")
	}
}

// An undo clone taken after the zones were answered keeps the answer, and
// answering on the clone leaves the live prompt asking (the frame is
// replaced, never written through).
func TestCloneCarriesTheZoneAnswer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Wanted")
	gyCard(me, "Wanted")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: named("Wanted"), Dest: ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	live := openSearch(g, me.ID)
	cl := g.Clone()
	c := openSearch(cl, me.ID)
	answerZones(t, cl, c, SearchZoneGraveyard)
	if !live.SearchZonesPending() || !openSearch(g, me.ID).SearchZonesPending() {
		t.Error("answering on the clone answered the live prompt")
	}
	cl2 := cl.Clone()
	if c2 := openSearch(cl2, me.ID); c2 == nil || c2.SearchZonesPending() || len(c2.SearchCards) != 1 {
		t.Errorf("a clone of an answered prompt must stay answered: %+v", c2)
	}
}
