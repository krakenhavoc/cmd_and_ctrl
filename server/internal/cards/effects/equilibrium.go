package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Equilibrium — Enchantment {1}{U}{U}:
//
//	"Whenever you cast a creature spell, you may pay {1}. If you do, return
//	 target creature to its owner's hand."
//
// A cast trigger whose target (any creature, yours, theirs, or the spell's
// future body once it is on the battlefield) is chosen as the trigger goes on
// the stack (CR 603.3d); the payment is made as it resolves, and the bounce
// happens only when the mana is really paid (Cradle of Vitality's shape).
// A target that has left by then is not returned (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "cbc052e7-817f-43b0-b8e3-96b1644d798b",
		Name:         "Equilibrium",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(Creature(), "Equilibrium — you may pay {1}; if you do, return target creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Equilibrium — pay {1} to return the target creature to its owner's hand?",
						Then:     equilibriumPaid,
					}.Apply(NewContext(g, item))
				}),
				TargetCreature("target creature")),
		},
	})
}

var equilibriumPaid = PayUnlessBranches("pay-unless/equilibrium-bounce",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}, nil)
