package game

import (
	"testing"

	"github.com/google/uuid"
)

// paused_entry_continuation_test.go — #478, the engine half. The
// catalog half (the issue's Kismet + Thalia repro, the exile return,
// the reanimation, the undo) is
// cards/effects/paused_entry_continuations_test.go; what is pinned here
// is the contract `entryTail` signs, which is the same one
// `zoneRoute.then` signs: the continuation runs from EVERY terminal
// outcome, not only from the landing, or a caller sequenced through it
// waits forever.

// TestCancelledEntryStillFinishesTheSearch is the outcome with no
// prompt in it and the one easiest to drop: the CR 614 window cancels
// the fetched permanent's entry outright (CR 614.10 with a null
// replacement). Nothing enters — and the search still happened, so
// EventSearchLibrary fires, the library is shuffled, and the caller's
// Then is told it found nothing.
func TestCancelledEntryStillFinishesTheSearch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := uuid.New()
	me.Library.PushTop(Card{
		InstanceID: target, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	thenRan := 0
	var found []uuid.UUID

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == target && ev.NewZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: "test: cancel the entry",
	})
	err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player:  me.ID,
		Pred:    func(c Card) bool { return c.Name == "Bears" },
		Dest:    ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Then:    func(_ *Game, ids []uuid.UUID) error { thenRan++; found = ids; return nil },
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("SearchLibraryThenForEffect: %v", err)
	}

	if thenRan != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", thenRan)
	}
	if len(found) != 0 {
		t.Errorf("found = %v, want nothing: the entry was cancelled", found)
	}
	if !me.Library.Contains(target) {
		t.Error("a cancelled entry leaves the card in the library")
	}
	if g.Battlefield.Contains(target) {
		t.Error("nothing entered")
	}
	fired := false
	for _, ev := range g.Events {
		if ev.Kind == EventSearchLibrary {
			fired = true
		}
	}
	if !fired {
		t.Error("the search still happened, so EventSearchLibrary fires and the library shuffles")
	}
}
