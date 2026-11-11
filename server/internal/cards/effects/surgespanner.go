package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surgespanner — Creature — Merfolk Wizard {2}{U}{U}, 2/2:
//
//	"Whenever this creature becomes tapped, you may pay {1}{U}. If you
//	 do, return target permanent to its owner's hand."
//
// The target is chosen when the trigger is put on the stack (CR 603.3d);
// the payment is asked as it resolves (MayPay), and only a payer who
// pays bounces the permanent. Becoming tapped is a tap by any cause: a
// {T} cost, an attack without vigilance, or an effect (ThisBecameTapped).
func init() {
	Register(Spec{
		OracleID:     "3f15349a-6ac6-4091-95a3-ae9f593e4650",
		Name:         "Surgespanner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Surgespanner — you may pay {1}{U} to return target permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{1}{U}",
							Question: "Surgespanner — pay {1}{U} to return the target permanent to its owner's hand?",
							Then:     cmdr2B05SurgespannerPaid,
						}.Apply(NewContext(g, item))
					}),
				TargetPermanent("target permanent"),
			),
		},
	})
}

var cmdr2B05SurgespannerPaid = PayUnlessBranches("pay-unless/cmdr2-b05-surgespanner-bounce",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return BounceToHand{Target: ctx.Item.Targets[0].ID}.Apply(ctx)
	}, nil)
