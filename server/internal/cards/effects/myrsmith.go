package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7fc3798a-3907-4f50-98a1-25b8e1977197",
		Name:         "Myrsmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Artifact(),
				"Myrsmith — you may pay {1} to create a 1/1 colorless Myr artifact creature token",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Myrsmith — pay {1} to create a 1/1 colorless Myr artifact creature token?",
						Then:     rprb25MyrsmithPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var rprb25MyrsmithPaid = PayUnlessBranches("pay-unless/s82-rpr-b25-myrsmith-myr",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 colorless Myr artifact"), N: 1}.Apply(ctx)
	}, nil)
