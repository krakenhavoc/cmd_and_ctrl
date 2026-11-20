package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emitDamageEventForTest emits the damage event the combat damage step
// would, from `source` to `target`.
func EmitDamageEventForTest(g *game.Game, source, target uuid.UUID, amount int, combat bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Source: source, Target: target,
			Amount: amount, Combat: combat,
		})
	})
}
