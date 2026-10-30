package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultron, Unlimited — Legendary Artifact Creature — Robot Villain {1}{U}{B},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Ultron attacks, he connives.
//	 Whenever a creature you control connives, you may pay {1}. If you do,
//	 create a 2/2 colorless Robot Villain artifact creature token."
//
// The attack trigger is Atlantean Skirmisher's sourceConnives. The second
// trigger watches EventConnived (CR 701.50f) like Iron Monger, so it sees
// Ultron's own connive as well as any other creature's, one trigger per
// conniving creature. The {1} is asked as the ability resolves (MayPay) and
// the token is made only when it was paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "436501ca-0998-43fc-975f-87c2d34caab3",
		Name:            "Ultron, Unlimited",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ultron, Unlimited — he connives", Do(Instruction(sourceConnives))),
			WheneverACreatureYouControlConnives("Ultron, Unlimited — you may pay {1} to create a 2/2 Robot Villain token",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Ultron, Unlimited — pay {1} to create a 2/2 colorless Robot Villain artifact creature token?",
						Then:     ultronUnlimitedPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// ultronUnlimitedPaid is the paid half, keyed: the Robot Villain token.
var ultronUnlimitedPaid = PayUnlessBranches("pay-unless/ultron-unlimited-robot-villain",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Template: TokenCard("2/2 colorless Robot Villain artifact"), N: 1}.Apply(ctx)
	}, nil)
