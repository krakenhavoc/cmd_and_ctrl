package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning Cloud — Enchantment {3}{R} (Visions, S82 OLD batch 786, #5631):
//
//	"Whenever a player casts a red spell, you may pay {R}. If you do, this
//	 enchantment deals 1 damage to any target."
//
// Quiet Contemplation's targeted trigger with a MayPay: any player's red
// spell triggers it (the colour is read through the enchantment's own text,
// CR 612.2), the any-target is chosen as the trigger goes on the stack
// (CR 603.3d), the {R} is paid as it resolves, and the damage is dealt only
// if it was paid and the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.Targeting(
		effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("R")),
			"Lightning Cloud — you may pay {R}: 1 damage to any target",
			func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{R}",
					Question: "Lightning Cloud — pay {R} to deal 1 damage to any target?",
					Then:     lightningCloudPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		effects.TargetAny())
	effects.Register(effects.Spec{
		OracleID:     "4a574140-0657-4f49-b497-331447f17b29",
		Name:         "Lightning Cloud",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// lightningCloudPaid is the paid half: 1 damage from the enchantment to its
// target, if still legal. An on-disk key: never renamed.
var lightningCloudPaid = effects.PayUnlessBranches("pay-unless/s82-old-b786-lightning-cloud-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: 1}.Apply(ctx)
	}, nil)
