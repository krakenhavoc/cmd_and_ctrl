package game

import (
	"testing"

	"github.com/google/uuid"
)

// library_shuffle_test.go — #6215: EventLibraryShuffled. CR 701.24a:
// "To shuffle a library or a face-down pile of cards, randomize the
// cards within it so that no player knows their order."

func shuffleEvents(g *Game, from int) []Event {
	var out []Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventLibraryShuffled {
			out = append(out, ev)
		}
	}
	return out
}

// The opening hands' shuffles are no spell or ability's: nothing is
// announced for them, and nothing for a mulligan.
func TestOpeningAndMulliganShufflesEmitNothing(t *testing.T) {
	g := newActiveGameMulligansOpen(t, 2)
	if n := len(shuffleEvents(g, 0)); n != 0 {
		t.Fatalf("opening shuffles announced %d events", n)
	}
	start := len(g.Events)
	if err := g.Mulligan(g.Seats[0].ID, 7); err != nil {
		t.Fatalf("Mulligan: %v", err)
	}
	if n := len(shuffleEvents(g, start)); n != 0 {
		t.Errorf("a mulligan announced %d shuffle events", n)
	}
}

// "Shuffle your library" from an effect is announced as an effect's,
// for that player; the sandbox's manual shuffle as an action.
func TestShuffleLibraryForEffectAnnouncesTheShuffle(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	start := len(g.Events)
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(me) })
	evs := shuffleEvents(g, start)
	if len(evs) != 1 || evs[0].Actor != me || evs[0].Label != ShuffleByEffect {
		t.Fatalf("events = %+v, want one effect shuffle of my library", evs)
	}
	start = len(g.Events)
	if err := g.ShuffleLibrary(me); err != nil {
		t.Fatal(err)
	}
	evs = shuffleEvents(g, start)
	if len(evs) != 1 || evs[0].Label != ShuffleByAction {
		t.Fatalf("manual shuffle events = %+v, want one action shuffle", evs)
	}
}

// CR 701.24e: a library of zero cards can still be shuffled, and a
// shuffle trigger fires.
func TestShufflingAnEmptyLibraryIsAnnounced(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		_ = g.ShuffleLibraryForEffect(me.ID)
	})
	if n := len(shuffleEvents(g, 0)); n != 1 {
		t.Fatalf("empty-library shuffle events = %d, want 1", n)
	}
}

// A fetch-and-shuffle announces the search and then the shuffle, of the
// searched library; a search that doesn't shuffle announces none.
func TestSearchWithShuffleAnnouncesTheShuffle(t *testing.T) {
	for _, shuffle := range []bool{true, false} {
		g := newActiveGame(t)
		a := g.Seats[0]
		basic := Card{InstanceID: uuid.New(), Name: "Wastes", TypeLine: "Basic Land", Owner: a.ID, Controller: a.ID}
		a.Library.PushTop(basic)
		start := len(g.Events)
		resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
			return g.SearchLibraryForEffect(a.ID, func(c Card) bool { return c.Name == "Wastes" }, ZoneHand, 1, false, shuffle)
		})
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == a.ID {
				if err := g.ResolveSearchLibrary(c.ID, a.ID, []uuid.UUID{basic.InstanceID}); err != nil {
					t.Fatalf("ResolveSearchLibrary: %v", err)
				}
				break
			}
		}
		evs := shuffleEvents(g, start)
		if shuffle && (len(evs) != 1 || evs[0].Actor != a.ID || evs[0].Label != ShuffleByEffect) {
			t.Errorf("shuffle: events = %+v, want one effect shuffle", evs)
		}
		if !shuffle && len(evs) != 0 {
			t.Errorf("no shuffle: events = %+v, want none", evs)
		}
	}
}
