package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dnDesignation reads the designation without the lock (tests are
// single-goroutine).
func DnDesignation(g *game.Game) game.DayNightDesignation { return g.DayNightDesignation() }

func CountKind(g *game.Game, kind game.EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}
