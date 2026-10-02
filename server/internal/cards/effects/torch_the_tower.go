package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "10d33e95-3e5d-447e-ba4a-acd3c33b4045",
		Name:          "Torch the Tower",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			bargained := g2WasBargained(ctx)
			amount := 2
			if bargained {
				amount = 3
			}
			if err := DealDamageThen(ctx, id, amount, ExileIfDealtDamageWouldDie(item, true)); err != nil {
				return err
			}
			if !bargained {
				return nil
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
