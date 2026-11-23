package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// playerCounterFor reads a seat's counter of the named kind.
func PlayerCounterFor(g *game.Game, playerID uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				n = p.Counters[kind]
			}
		}
	})
	return n
}
