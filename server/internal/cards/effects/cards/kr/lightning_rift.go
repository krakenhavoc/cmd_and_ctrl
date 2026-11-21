package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Rift — Enchantment {1}{R} (Dominaria Remastered, #5948):
//
//	"Whenever a player cycles a card, you may pay {1}. If you do, this
//	 enchantment deals 2 damage to any target."
//
// A cycle trigger of any player with a target chosen as it goes on the
// stack; the payment is asked as it resolves, and "if you do" is the paid
// branch, which deals the damage from the enchantment (last-known
// information) to that target if it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac3fb7ef-709c-40be-a75b-1b438708388d",
		Name:         "Lightning Rift",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCycle, effects.AnyPlayer, "Lightning Rift — you may pay {1}; if you do, it deals 2 damage to any target",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayPay{
							Chooser:  item.Controller,
							Cost:     "{1}",
							Question: "Lightning Rift — pay {1} to deal 2 damage to the target?",
							Then:     lightningRiftPaid,
						}.Apply(effects.NewContext(g, item))
					}),
				effects.TargetAny(),
			),
		},
	})
}

var lightningRiftPaid = effects.PayUnlessBranches("pay-unless/lightning-rift-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.DealDamageToTheTarget(2)(ctx.Game, ctx.Item)
	}, nil)
