package game

import (
	"testing"

	"github.com/google/uuid"
)

// search_top_n_test.go — #6090: Aven Mindcensor's "If an opponent would
// search a library, that player searches the top four cards of that
// library instead" (search_top_n.go).

const testMindcensorKey = "test-mindcensor"

func withSearchTopN(t *testing.T, n int) {
	t.Helper()
	prev := CatalogOpponentsSearchTopN
	CatalogOpponentsSearchTopN = func(key string) int {
		if key == testMindcensorKey {
			return n
		}
		return 0
	}
	t.Cleanup(func() { CatalogOpponentsSearchTopN = prev })
}

// An opponent's search sees only the top four: a match lower down is not a
// candidate. The whole library is still shuffled and the search still
// happened.
func TestOpponentSearchesOnlyTheTopN(t *testing.T) {
	withSearchTopN(t, 4)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldForTest(g, opp.ID, "Mindcensor", "Creature — Bird Wizard", testMindcensorKey)
	// Bottom first: the deep Forest is fifth from the top.
	ids := searchLibraryOf(me, "Island", "Forest", "Swamp", "Plains", "Forest", "Mountain", "Island")
	deep, high := ids[1], ids[4]

	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: func(c Card) bool { return c.Name == "Forest" },
		Dest: ZoneHand, Limit: 2, Shuffle: true,
	})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != high {
		t.Fatalf("candidates = %v, want only the Forest in the top four (%v)", c.SearchCards, high)
	}
	g.mu.Lock()
	deepLegal := g.SearchPickLegalLocked(c, []uuid.UUID{deep})
	g.mu.Unlock()
	if deepLegal {
		t.Error("the Forest below the top four can be taken")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{high}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !me.Hand.Contains(high) || me.Hand.Contains(deep) {
		t.Error("the wrong Forest moved")
	}
	if searchEvents(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", searchEvents(g))
	}
}

// The controller's own search, and every search once the permanent loses
// its key, look through the whole library.
func TestSearchTopNSparesItsControllerAndNeedsTheAbility(t *testing.T) {
	withSearchTopN(t, 4)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldForTest(g, me.ID, "Mindcensor", "Creature — Bird Wizard", testMindcensorKey)
	ids := searchLibraryOf(me, "Forest", "Island", "Island", "Island", "Island", "Island")
	runSearch(t, g, SearchLibrarySpec{
		Player: me.ID, Pred: func(c Card) bool { return c.Name == "Forest" }, Dest: ZoneHand, Limit: 1,
	})
	c := openSearch(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("the controller's own search was narrowed: %+v", c)
	}
	if g.searchTopNLocked(opp.ID) != 4 {
		t.Error("an opponent is not narrowed")
	}
	CatalogOpponentsSearchTopN = nil
	if g.searchTopNLocked(opp.ID) != 0 {
		t.Error("a nil hook still narrows")
	}
}

// A library shorter than N is searched whole.
func TestSearchTopNShortLibrary(t *testing.T) {
	withSearchTopN(t, 4)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldForTest(g, opp.ID, "Mindcensor", "Creature — Bird Wizard", testMindcensorKey)
	lib := []Card{{Name: "a"}, {Name: "b"}}
	g.mu.Lock()
	got := g.searchableLibraryCards(me.ID, lib)
	g.mu.Unlock()
	if len(got) != 2 {
		t.Errorf("a two-card library searched %d cards, want 2", len(got))
	}
}
