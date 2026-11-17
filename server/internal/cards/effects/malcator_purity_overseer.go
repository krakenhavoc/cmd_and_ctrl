package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malcator, Purity Overseer — Legendary Creature — Phyrexian Elephant
// Wizard {1}{W}{U}, 1/1 (Phyrexia: All Will Be One):
//
//	"When Malcator enters, create a 3/3 colorless Phyrexian Golem artifact
//	 creature token.
//	 At the beginning of your end step, if three or more artifacts entered
//	 the battlefield under your control this turn, create a 3/3 colorless
//	 Phyrexian Golem artifact creature token."
//
// The end-step trigger is an intervening "if" (CR 603.4) over the per-turn
// entry record, which counts every artifact that entered under your
// control this turn, tokens included, even one that has since left.
//
// No simplification.
func init() {
	golem := Do(CreateToken{Template: TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})
	Register(Spec{
		OracleID:     "0d1c6a82-943a-46bd-97ed-6055226ef2ab",
		Name:         "Malcator, Purity Overseer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Malcator, Purity Overseer — create a 3/3 Phyrexian Golem", golem),
			InterveningIf(CondBoard{
				Name: "three or more artifacts entered the battlefield under your control this turn",
				Test: func(ctx *Context) bool {
					return ctx.Game.EnteredWithCardTypeThisTurn(ctx.Controller(), "artifact") >= 3
				},
			}, AtYourEndStep("Malcator, Purity Overseer — create a 3/3 Phyrexian Golem", golem)),
		},
	})
}
