package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigil for the Lost — Enchantment {3}{W} (Scars of Mirrodin, S82 OLD batch
// 472, #5577):
//
//	"Whenever a creature you control dies, you may pay {X}. If you do, you gain
//	 X life."
//
// Well of Lost Dreams' "you may pay {X}" (MayPayX): X is chosen as the trigger
// resolves and paid through the ordinary payment prompt, and paying {0} gains
// nothing. One trigger per creature that dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87a4caa4-cb08-4a0c-b57a-d6d8474b1f5e",
		Name:         "Vigil for the Lost",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Vigil for the Lost — pay {X} to gain X life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Cost:  "{X}",
						Label: "Vigil for the Lost",
						Buys:  "gain X life",
						Goal:  effects.XAsHighAsYouCan,
						Unit:  game.PayAmountOther,
						Then:  vigilForTheLostPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var vigilForTheLostPaid = effects.PayUnlessBranches("pay-unless/s82-old-b472-vigil-for-the-lost-life",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.GainLife{Player: ctx.Controller(), Amount: r.Number}.Apply(ctx)
	}, nil)
