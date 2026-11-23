package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch the Tower — Instant {R}:
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Torch the Tower deals 2 damage to target creature or planeswalker.
//	 If this spell was bargained, instead it deals 3 damage to that
//	 permanent and you scry 1.
//	 If a permanent dealt damage by Torch the Tower would die this turn,
//	 exile it instead."
//
// Bargain is an optional additional cost (g2Bargain), announced and
// paid with the cast like a sacrifice kicker; "bargained" is read back
// from the payment record at resolution. "A permanent dealt damage by
// Torch the Tower" is registered from the damage's continuation (ADR
// 0108 §1 decision 3), on any permanent that took more than 0 damage.
// The scry comes after the damage, in printed order.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "10d33e95-3e5d-447e-ba4a-acd3c33b4045",
		Name:          "Torch the Tower",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OptionalCosts: []game.AdditionalCost{effects.G2Bargain()},
		// The scry waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				amount := 2
				if effects.G2WasBargained(ctx) {
					amount = 3
				}
				return effects.DealDamageThen(ctx, id, amount, effects.ExileIfDealtDamageWouldDie(ctx.Item, true))
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok || !effects.G2WasBargained(ctx) {
					return nil
				}
				return effects.Scry{N: 1}.Apply(ctx)
			}),
		),
	})
}
