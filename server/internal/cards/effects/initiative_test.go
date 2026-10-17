package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative_test.go pins ADR 0203 (#4096): the initiative (CR 726),
// venturing into the Undercity (CR 309, CR 701.49) and the room
// abilities.

// initTake gives `player` the initiative through the card-side
// primitive, under the lock a resolving effect holds.
func initTake(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (TakeTheInitiative{Player: player}).Apply(&Context{Game: g}); err != nil {
			t.Errorf("TakeTheInitiative: %v", err)
		}
	})
}

// initRoom is `player`'s venture marker room, "" when in no dungeon.
func initRoom(g *game.Game, player uuid.UUID) string {
	var room string
	g.WithWriteLock(func() {
		if d := game.DungeonOf(g, player); d != nil {
			room = d.Room
		}
	})
	return room
}

// initPending is the open prompt of this kind addressed to `chooser`,
// or nil.
func initPending(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// initPickRoom answers the open venture branch prompt with the room
// whose key is `room`.
func initPickRoom(t *testing.T, g *game.Game, chooser uuid.UUID, room string) {
	t.Helper()
	c := initPending(g, game.PendingChoiceOptionPick, chooser)
	if c == nil {
		t.Fatalf("no venture prompt open for %v (pending %d)", chooser, len(g.PendingChoices))
	}
	for i, o := range c.PickOptions {
		if o.Room == room {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick %s: %v", room, err)
			}
			return
		}
	}
	t.Fatalf("room %q not offered: %+v", room, c.PickOptions)
}

// TestTakingTheInitiativeVenturesIntoSecretEntrance is the headline: a
// card's "you take the initiative" hands over the designation, and the
// inherent "whenever a player takes the initiative" trigger enters the
// Undercity's first room, whose ability searches for a basic land.
func TestTakingTheInitiativeVenturesIntoSecretEntrance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	initTake(t, g, me.ID)
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want %v", g.Initiative, me.ID)
	}
	if room := initRoom(g, me.ID); room != "" {
		t.Fatalf("ventured before the trigger resolved: %q", room)
	}
	monSettle(t, g)
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	c := initPending(g, game.PendingChoiceSearchLibrary, me.ID)
	if c == nil {
		t.Fatalf("Secret Entrance asked for no search (pending %d)", len(g.PendingChoices))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if z := g.FindCardZoneForEffect(forest); z == nil || z.Kind != game.ZoneHand {
		t.Fatalf("the Forest is not in hand")
	}
}
