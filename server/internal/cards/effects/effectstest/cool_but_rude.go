package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponentLives snapshots every opponent's life total.
func OpponentLives(g *game.Game, me uuid.UUID) map[uuid.UUID]int {
	out := make(map[uuid.UUID]int)
	for _, p := range g.Seats {
		if p.ID != me {
			out[p.ID] = p.Life
		}
	}
	return out
}
