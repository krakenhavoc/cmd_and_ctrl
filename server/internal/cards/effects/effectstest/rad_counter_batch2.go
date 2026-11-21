package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radTriggersFrom counts the waiting or stacked triggered items of one
// source.
func RadTriggersFrom(g *game.Game, source uuid.UUID) int {
	return len(TriggerAmountsFrom(g, source))
}
