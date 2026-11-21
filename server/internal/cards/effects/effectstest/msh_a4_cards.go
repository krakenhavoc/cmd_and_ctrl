package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MshA4Ref(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// mshA4Damage deals noncombat damage to a creature as one damage
// instruction and lets the enrage trigger, which waits for the batch, fire.
func MshA4Damage(g *game.Game, source, target uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, target, n) })
	g.RunStateChecksForTest()
}
