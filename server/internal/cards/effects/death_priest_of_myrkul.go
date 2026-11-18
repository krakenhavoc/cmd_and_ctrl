package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death-Priest of Myrkul — Creature — Tiefling Cleric {2}{B}{B}, 2/2
// (Adventures in the Forgotten Realms, S82):
//
//	"Skeletons, Vampires, and Zombies you control get +1/+1.
//	 At the beginning of your end step, if a creature died this turn, you may
//	 pay {1}. If you do, create a 1/1 black Skeleton creature token."
//
// The anthem is the tribal anthem over three creature types. The end-step
// trigger is Bulette's intervening "if" with Antagonist Incorporator's
// optional {1} payment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af231505-bd2e-4499-9fd5-f761dde2ecef",
		Name:         "Death-Priest of Myrkul",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Skeleton", "Vampire", "Zombie"}, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				AtYourEndStep("Death-Priest of Myrkul — you may pay {1} to create a 1/1 black Skeleton token",
					func(g *game.Game, item *game.StackItem) error {
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{1}",
							Question: "Pay {1} to create a 1/1 black Skeleton creature token?",
							Then:     deathPriestOfMyrkulPaid,
						}.Apply(NewContext(g, item))
					})),
		},
	})
}

var deathPriestOfMyrkulPaid = PayUnlessBranches("pay-unless/death-priest-of-myrkul-skeleton",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 black Skeleton"), N: 1}.Apply(ctx)
	}, nil)
