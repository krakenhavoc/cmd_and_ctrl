package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kohQueuedTrigger is a triggered item from `source` waiting to be put on
// the stack (CR 603.3) or already on it.
func KohQueuedTrigger(g *game.Game, source uuid.UUID) *game.StackItem {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	return TriggerOnStack(g, source)
}
