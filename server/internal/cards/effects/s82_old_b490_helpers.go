package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b490_helpers.go — Worldwake batch 490 (S82 older sets, #5582).

// S82OldB490DamageToTargetThenReturnGrantor is "It deals N damage to any
// target. Return <the Equipment> to its owner's hand" (Razor Boomerang): the
// equipped creature deals the damage; the Equipment, unattached as a cost, is
// returned only if it is still on the battlefield as the ability resolves.
func S82OldB490DamageToTargetThenReturnGrantor(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if err := DealDamageToTheTarget(n)(g, item); err != nil {
			return err
		}
		ctx := NewContext(g, item)
		equipment := ctx.Grantor()
		if equipment == uuid.Nil || !OnBattlefield(g, equipment) {
			return nil
		}
		return BounceToHand{Target: equipment}.Apply(ctx)
	}
}
