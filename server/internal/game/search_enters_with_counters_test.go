package game

import "testing"

// search_enters_with_counters_test.go — #2098: a library search that puts
// a card onto the battlefield can seed counters on the entry event
// (SearchLibrarySpec.EntersWithCounters), the way a token's do (#762).

func searchWithCounters(t *testing.T, g *Game, spec SearchLibrarySpec) {
	t.Helper()
	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(spec)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}
}

// The counter is on the permanent, it was placed BEFORE EventETB (so an
// enters trigger reading its counters sees it), and the spec's own map is
// not consumed by the entry.
func TestSearchedCardEntersWithCountersBeforeETB(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := topOfLibraryFor(me, "Bear", "Creature — Bear")
	counters := map[string]int{"+1/+1": 1}
	before := len(g.Events)

	searchWithCounters(t, g, SearchLibrarySpec{
		Player:             me.ID,
		Pred:               func(c Card) bool { return c.InstanceID == id },
		Dest:               ZoneBattlefield,
		Limit:              1,
		EntersWithCounters: counters,
	})

	c := battlefieldCardForTest(t, g, id)
	if got := c.Counters["+1/+1"]; got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	placed, etb := -1, -1
	for i, ev := range g.Events[before:] {
		switch {
		case ev.Kind == EventCounterPlaced && ev.Target == id && placed < 0:
			placed = i
		case ev.Kind == EventETB && ev.CardID == id:
			etb = i
		}
	}
	if etb < 0 {
		t.Fatal("no EventETB for the fetched card")
	}
	if placed < 0 || placed > etb {
		t.Errorf("counter event at %d, ETB at %d: the counter must be placed as it enters", placed, etb)
	}
	if counters["+1/+1"] != 1 || len(counters) != 1 {
		t.Errorf("the spec's map was mutated: %v", counters)
	}
}

// No option, no counters: the default search is untouched.
func TestSearchedCardWithoutOptionEntersBare(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := topOfLibraryFor(me, "Bear", "Creature — Bear")
	searchWithCounters(t, g, SearchLibrarySpec{
		Player: me.ID,
		Pred:   func(c Card) bool { return c.InstanceID == id },
		Dest:   ZoneBattlefield,
		Limit:  1,
	})
	if got := battlefieldCardForTest(t, g, id).Counters["+1/+1"]; got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
}
