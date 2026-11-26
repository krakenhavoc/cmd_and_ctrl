package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblinslide — Enchantment {2}{R} (Khans of Tarkir):
//
//	"Whenever you cast a noncreature spell, you may pay {1}. If you do, create
//	 a 1/1 red Goblin creature token with haste."
//
// Moku's cast trigger with the {1} paid as the trigger resolves (CR 118.12,
// the MayPay prompt); declining does nothing. The token is created for the
// controller of the trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4885354f-0b77-47b6-b8f7-00753804436f",
		Name:         "Goblinslide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Goblinslide — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Pay {1} to create a 1/1 red Goblin creature token with haste?",
					Then:     goblinslidePaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// goblinslidePaid is the paid half: the hasty Goblin. An on-disk key.
var goblinslidePaid = effects.PayUnlessBranches("pay-unless/s82-old-b351-goblinslide-goblin",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin with haste"), N: 1}.Apply(ctx)
	}, nil)
