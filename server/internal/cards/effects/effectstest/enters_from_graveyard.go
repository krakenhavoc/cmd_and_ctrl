package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lastETB returns the most recent EventETB for id.
func EfgLastETB(t *testing.T, g *game.Game, id uuid.UUID) game.Event {
	t.Helper()
	var found game.Event
	ok := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventETB && ev.CardID == id {
				found, ok = ev, true
			}
		}
	})
	if !ok {
		t.Fatalf("no EventETB for %s", id)
	}
	return found
}
