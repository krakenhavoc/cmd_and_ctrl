package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counteredInLog reports whether the log records `stackID` being
// countered.
func CounteredInLog(g *game.Game, stackID uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.Target == stackID {
			return true
		}
	}
	return false
}
