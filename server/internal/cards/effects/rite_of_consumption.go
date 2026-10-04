package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of Consumption — Sorcery {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Rite of Consumption deals damage equal to the sacrificed creature's
//	 power to target player or planeswalker. You gain life equal to the
//	 damage dealt this way."
//
// The damage is the sacrificed creature's last-known power (CR 608.2h),
// read off the payment record (ADR 0113 §1). The life is the damage
// actually DEALT, after prevention and doubling, through the same
// damage-then-total walk Creeping Bloodsucker uses: a prevented point
// gains nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "327121a1-f193-44c6-a834-802095abec84",
		Name:           "Rite of Consumption",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        targetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := ctx.SacrificedPower()
			if len(item.Targets) == 0 || n <= 0 {
				return nil
			}
			controller, source := ctx.Controller(), ctx.Source()
			return ctx.Game.DealDamageEachThenForEffect(source, []uuid.UUID{item.Targets[0].ID}, n,
				func(g *game.Game, dealt int) error {
					if dealt <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(source, controller, dealt)
				})
		},
	})
}
