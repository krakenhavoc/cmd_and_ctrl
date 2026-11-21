package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moku, Meandering Drummer — Legendary Creature — Human Bard Ally {1}{R},
// 2/2 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever you cast a noncreature spell, you may pay {1}. If you do,
//	 Moku gets +2/+1 and creatures you control gain haste until end of
//	 turn."
//
// The {1} is paid as the trigger resolves (CR 118.12), through the MayPay
// prompt (Era of Innovation's shape); declining does nothing. The paid
// branch pumps Moku only while it is still the permanent the trigger came
// from (CR 400.7), and gives haste to the creatures you control as it
// resolves (CR 611.2c), Moku among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6e1b1ff-e4e5-4814-9a20-2d527014463f",
		Name:         "Moku, Meandering Drummer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Moku, Meandering Drummer — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Pay {1}: Moku gets +2/+1 and creatures you control gain haste?",
					Then:     avB30aMokuPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// avB30aMokuPaid is the paid half, keyed (#4264).
var avB30aMokuPaid = effects.PayUnlessBranches("pay-unless/avatar-b30-moku-pump-and-haste",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Moku — +2/+1"}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"haste"},
			Label: "Moku — creatures you control gain haste"}.Apply(ctx)
	}, nil)
