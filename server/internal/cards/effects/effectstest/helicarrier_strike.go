package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// helicarrierAttacker stamps a creature's AttackingTarget directly —
// the target predicate only cares that the field is set, not that a
// full combat declaration produced it.
func HelicarrierAttacker(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
				return
			}
		}
	})
}
