package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CommitCrimeForMinerTest(g *game.Game, actor uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, CrimeCommitted: true, Actor: actor})
	})
}
