package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roll_events.go — die-roll event helpers shared across test packages.

// RollEventsOf lists the EventRollDie events in the log.
func RollEventsOf(g *game.Game) []game.Event {
	var out []game.Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventRollDie {
				out = append(out, ev)
			}
		}
	})
	return out
}
