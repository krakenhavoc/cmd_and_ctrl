package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// intelCards lists the player's exiled cards that carry an intel counter.
func IntelCards(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, e := range g.Exile.Cards {
		if e.Owner == player && e.Counters["intel"] > 0 {
			out = append(out, e.InstanceID)
		}
	}
	return out
}
