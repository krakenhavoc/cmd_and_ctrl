package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func BeginUpkeep(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: who}) })
	PassPriorityAroundTable(t, g)
}
