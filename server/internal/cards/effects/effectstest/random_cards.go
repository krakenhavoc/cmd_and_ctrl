package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RandomEvents(g *game.Game, kind game.EventKind) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}
func RandomTokenCount(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name {
			n++
		}
	}
	return n
}
