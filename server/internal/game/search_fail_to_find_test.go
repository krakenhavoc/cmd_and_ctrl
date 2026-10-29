package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// search_fail_to_find_test.go — #3053 (and #3081): CR 701.23b/d.
//
// A search of a hidden zone for a card with a stated quality may find
// nothing, even when exactly one card matches; "up to N" with a stated
// quality may take 0..N. A search simply for a quantity ("a card",
// "three cards") must find that many. A public zone's search must find
// a valid card when one is present. Whatever is found, the rest of the
// instruction (shuffle, EventSearchLibrary, Then) still happens.

func TestSearchFloorTable(t *testing.T) {
	cases := []struct {
		name                   string
		hidden, quality, fewer bool
		limit, matches, want   int
	}{
		{"stated quality, one match", true, true, false, 1, 1, 0},
		{"stated quality, many matches", true, true, false, 1, 9, 0},
		{"stated quality, up to two of two", true, true, false, 2, 2, 0},
		{"unconditional, one card, one in library", true, false, false, 1, 1, 1},
		{"unconditional, one card, many in library", true, false, false, 1, 40, 1},
		{"unconditional, three cards, many", true, false, false, 3, 40, 3},
		{"unconditional, three cards, two in library", true, false, false, 3, 2, 2},
		{"unconditional, empty library", true, false, false, 1, 0, 0},
		{"unconditional, up to three", true, false, true, 3, 40, 0},
		{"you may search, unconditional", true, false, true, 1, 40, 0},
		// The rule for a zone everybody can see: no fail to find, however
		// the card is worded, unless the clause itself says "up to".
		{"public zone, stated quality, one match", false, true, false, 1, 1, 1},
		{"public zone, stated quality, many matches", false, true, false, 1, 9, 1},
		{"public zone, two cards, one present", false, true, false, 2, 1, 1},
		{"public zone, none present", false, true, false, 1, 0, 0},
		{"public zone, up to two", false, true, true, 2, 5, 0},
	}
	for _, c := range cases {
		if got := searchFloor(c.hidden, c.quality, c.fewer, c.limit, c.matches); got != c.want {
			t.Errorf("%s: floor = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestSpecFloorReadsTheSpec(t *testing.T) {
	isLand := func(c Card) bool { return c.Name == "Forest" }
	if got := (SearchLibrarySpec{Pred: isLand, Limit: 1}).floor(1); got != 0 {
		t.Errorf("a predicate is a stated quality: floor = %d, want 0", got)
	}
	if got := (SearchLibrarySpec{Rule: testSearchRuleAtMost.WithN(2), Limit: 2}).floor(2); got != 0 {
		t.Errorf("a set rule is a stated quality: floor = %d, want 0", got)
	}
	if got := (SearchLibrarySpec{Limit: 1}).floor(5); got != 1 {
		t.Errorf("no predicate is an unconditional search: floor = %d, want 1", got)
	}
	if got := (SearchLibrarySpec{Limit: 3, UpTo: true}).floor(5); got != 0 {
		t.Errorf("UpTo: floor = %d, want 0", got)
	}
	if got := (SearchLibrarySpec{Limit: 1, Optional: true}).floor(5); got != 0 {
		t.Errorf("Optional: floor = %d, want 0", got)
	}
	if got := (SearchLibrarySpec{Unbounded: true}).floor(5); got != 0 {
		t.Errorf("Unbounded: floor = %d, want 0", got)
	}
}

// searchLibraryOf replaces the library with exactly `names`, bottom first, all
// plain "Land" cards, and returns the IDs in the same order.
func searchLibraryOf(p *Player, names ...string) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(names))
	for _, n := range names {
		id := uuid.New()
		p.Library.PushTop(Card{InstanceID: id, Name: n, TypeLine: "Basic Land — " + n, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func openSearch(g *Game, who uuid.UUID) *PendingChoice {
	var out *PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceSearchLibrary && c.Chooser == who {
				out = c
			}
		}
	})
	return out
}

func runSearch(t *testing.T, g *Game, spec SearchLibrarySpec) {
	t.Helper()
	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(spec)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}
}

func searchEvents(g *Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventSearchLibrary {
				n++
			}
		}
	})
	return n
}

// wasShuffled: the shuffle is what un-knows the library, so a library
// the searcher saw into and that holds no knowers was shuffled.
func libraryKnownTo(p *Player, who uuid.UUID) int {
	n := 0
	for _, c := range p.Library.Cards {
		if c.KnownBy[who] {
			n++
		}
	}
	return n
}

// The case the issue is about: ONE match for a stated quality. It used
// to be taken without asking. It asks; failing to find is honoured; the
// library is still shuffled, the event still fires, Then still runs.
func TestOneMatchForAStatedQualityOffersFailToFind(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Island", "Swamp", "Mountain")
	swamp := ids[1]
	thenRan, thenFound := 0, -1

	runSearch(t, g, SearchLibrarySpec{
		Player:   me.ID,
		Pred:     func(c Card) bool { return c.Name == "Swamp" },
		Dest:     ZoneHand,
		Limit:    1,
		Shuffle:  true,
		ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { thenRan++; thenFound = len(found); return nil }),
	})

	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("a one-match search for a stated quality did not prompt (CR 701.23b)")
	}
	if c.SearchMin != 0 || c.SearchMax != 1 {
		t.Fatalf("SearchMin/Max = %d/%d, want 0/1", c.SearchMin, c.SearchMax)
	}
	if me.Hand.Contains(swamp) {
		t.Fatal("the match moved before the searcher answered")
	}
	if libraryKnownTo(me, me.ID) == 0 {
		t.Fatal("the searcher should know the candidate while the prompt is open")
	}

	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("failing to find was refused: %v", err)
	}
	if !me.Library.Contains(swamp) || me.Hand.Contains(swamp) {
		t.Error("failing to find still took the card")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library size = %d, want 3", me.Library.Size())
	}
	if searchEvents(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: the search still happened", searchEvents(g))
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled after failing to find")
	}
	if thenRan != 1 || thenFound != 0 {
		t.Errorf("Then ran %d times with %d found, want once with none", thenRan, thenFound)
	}
}

func TestOneMatchForAStatedQualityCanBeTaken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Island", "Swamp", "Mountain")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: func(c Card) bool { return c.Name == "Swamp" },
		Dest: ZoneHand, Limit: 1, Shuffle: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the taken card is not in hand")
	}
}

// "Up to N" with a stated quality: take 0, 1 or N, whatever the library
// holds — here exactly N matches, the case that was taken whole.
func TestUpToNWithAStatedQualityMayTakeFewer(t *testing.T) {
	for _, take := range []int{0, 1, 2} {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := searchLibraryOf(me, "Forest", "Forest", "Island")
		runSearch(t, g, SearchLibrarySpec{
			Player: me.ID, Pred: func(c Card) bool { return c.Name == "Forest" },
			Dest: ZoneHand, Limit: 2, Shuffle: true,
		})
		c := openSearch(g, me.ID)
		if c == nil {
			t.Fatalf("take %d: no prompt although exactly Limit cards match", take)
		}
		if c.SearchMin != 0 || c.SearchMax != 2 {
			t.Fatalf("SearchMin/Max = %d/%d, want 0/2", c.SearchMin, c.SearchMax)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:take]); err != nil {
			t.Fatalf("take %d: %v", take, err)
		}
		held := 0
		for _, id := range ids[:2] {
			if me.Hand.Contains(id) {
				held++
			}
		}
		if held != take {
			t.Errorf("took %d of the matches, want %d", held, take)
		}
	}
}

// Demonic Tutor: no stated quality, so it must find. One card in the
// library: nothing to decide, taken unasked. Several: a prompt with a
// floor, and the empty answer is refused.
func TestAnUnconditionalSearchMustFind(t *testing.T) {
	t.Run("one card in the library is taken without asking", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := searchLibraryOf(me, "Island")
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true})
		if openSearch(g, me.ID) != nil {
			t.Fatal("a forced pick queued a prompt")
		}
		if !me.Hand.Contains(ids[0]) {
			t.Error("the only card was not found")
		}
		if searchEvents(g) != 1 {
			t.Errorf("EventSearchLibrary fired %d times, want 1", searchEvents(g))
		}
	})

	t.Run("an empty library finds nothing but still shuffles and fires", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me)
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true})
		// #4140: nothing to find is still asked, so the table cannot tell
		// an empty library from a full one by the absence of a prompt.
		c := openSearch(g, me.ID)
		if c == nil || len(c.SearchCards) != 0 {
			t.Fatalf("an empty library should still prompt, with no candidates: %+v", c)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
			t.Fatalf("finding nothing: %v", err)
		}
		if searchEvents(g) != 1 {
			t.Errorf("EventSearchLibrary fired %d times, want 1", searchEvents(g))
		}
	})

	t.Run("several cards: the empty answer is refused, a pick is accepted", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := searchLibraryOf(me, "Island", "Swamp", "Mountain")
		thenFound := -1
		runSearch(t, g, SearchLibrarySpec{
			Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true,
			ThenTail: searchHook(func(_ *Game, found []uuid.UUID) error { thenFound = len(found); return nil }),
		})
		c := openSearch(g, me.ID)
		if c == nil {
			t.Fatal("no prompt")
		}
		if c.SearchMin != 1 || c.SearchMax != 1 {
			t.Fatalf("SearchMin/Max = %d/%d, want 1/1", c.SearchMin, c.SearchMax)
		}
		err := g.ResolveSearchLibrary(c.ID, me.ID, nil)
		if !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("failing to find an unconditional search: err = %v, want ErrInvalidParam", err)
		}
		if openSearch(g, me.ID) == nil {
			t.Fatal("the refused answer dequeued the prompt")
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
			t.Fatalf("take: %v", err)
		}
		if !me.Hand.Contains(ids[2]) || thenFound != 1 {
			t.Errorf("the pick did not arrive (hand=%v, then found %d)", me.Hand.Contains(ids[2]), thenFound)
		}
	})

	t.Run("three cards means three", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := searchLibraryOf(me, "A", "B", "C", "D", "E")
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 3, Shuffle: true})
		c := openSearch(g, me.ID)
		if c == nil || c.SearchMin != 3 || c.SearchMax != 3 {
			t.Fatalf("prompt = %+v, want a 3..3 pick", c)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:2]); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("two of three: err = %v, want ErrInvalidParam", err)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:3]); err != nil {
			t.Fatalf("three of three: %v", err)
		}
	})

	t.Run("three cards from a library of two takes both", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := searchLibraryOf(me, "A", "B")
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 3, Shuffle: true})
		if openSearch(g, me.ID) != nil {
			t.Fatal("as many as possible needs no decision")
		}
		if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
			t.Error("not every card was found")
		}
	})

	t.Run("up to N with no quality may take fewer", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me, "A", "B", "C", "D")
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 3, UpTo: true, Shuffle: true})
		c := openSearch(g, me.ID)
		if c == nil || c.SearchMin != 0 || c.SearchMax != 3 {
			t.Fatalf("prompt = %+v, want a 0..3 pick", c)
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
			t.Fatalf("up to three, take none: %v", err)
		}
	})
}

// A rule that says "as many as possible" must not wedge the table when
// the candidates are gone: the floor shrinks with the library.
func TestTheFloorFallsWithTheCandidatesStillInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "A", "B", "C")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	for _, id := range ids {
		if err := g.MoveCardByID(ZoneRef{Kind: ZoneLibrary, Owner: me.ID},
			ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatalf("mill: %v", err)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("nothing is left to find, so finding nothing must be accepted: %v", err)
	}
}

func TestSearchFloorLockedFollowsTheChoice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "A", "B", "C")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 2, Shuffle: true})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	g.ReadSnapshot(func() {
		if got := g.SearchFloorLocked(c); got != 2 {
			t.Errorf("SearchFloorLocked = %d, want 2", got)
		}
	})
}
