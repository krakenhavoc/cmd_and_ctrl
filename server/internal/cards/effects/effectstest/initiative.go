package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative_test.go pins ADR 0203 (#4096): the initiative (CR 726),
// venturing into the Undercity (CR 309, CR 701.49) and the room
// abilities.

// initTake gives `player` the initiative through the card-side
// primitive, under the lock a resolving effect holds.
func InitTake(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.TakeTheInitiative{Player: player}).Apply(&effects.Context{Game: g}); err != nil {
			t.Errorf("TakeTheInitiative: %v", err)
		}
	})
}

// initRoom is `player`'s venture marker room, "" when in no dungeon.
func InitRoom(g *game.Game, player uuid.UUID) string {
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
func InitPending(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// initPickRoom answers the open venture branch prompt with the room
// whose key is `room`.
func InitPickRoom(t *testing.T, g *game.Game, chooser uuid.UUID, room string) {
	t.Helper()
	c := InitPending(g, game.PendingChoiceOptionPick, chooser)
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
