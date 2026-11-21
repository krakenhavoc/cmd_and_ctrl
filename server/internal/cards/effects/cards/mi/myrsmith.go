package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myrsmith — Creature — Human Artificer {1}{W}, 2/1:
//
//	"Whenever you cast an artifact spell, you may pay {1}. If you do,
//	 create a 1/1 colorless Myr artifact creature token."
//
// One cast trigger (WheneverYouCast over an artifact spell) whose
// resolution asks for the {1} (MayPay), Fathom Fleet Captain's shape; the
// token is made only once the payment has been made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fc3798a-3907-4f50-98a1-25b8e1977197",
		Name:         "Myrsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Artifact(),
				"Myrsmith — you may pay {1} to create a 1/1 colorless Myr artifact creature token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Myrsmith — pay {1} to create a 1/1 colorless Myr artifact creature token?",
						Then:     rprb25MyrsmithPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var rprb25MyrsmithPaid = effects.PayUnlessBranches("pay-unless/s82-rpr-b25-myrsmith-myr",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 colorless Myr artifact"), N: 1}.Apply(ctx)
	}, nil)
