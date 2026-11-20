package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feudkiller's Verdict — Kindred Sorcery — Giant {4}{W}{W} (Morningtide,
// #5592):
//
//	"You gain 10 life. Then if you have more life than an opponent, create
//	 a 5/5 white Giant Warrior creature token."
//
// Two steps in order (CR 608.2c): the life gain, then the comparison,
// made after the gain, against each opponent still in the game; more life
// than any one of them is enough.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dacd89b-5983-4905-9f6c-43aa6a12149d",
		Name:         "Feudkiller's Verdict",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: ResolveSteps(
			GainLife{Amount: 10},
			Instruction(func(ctx *Context) error {
				if !morMoreLifeThanAnOpponent(ctx.Game, ctx.Controller()) {
					return nil
				}
				return CreateToken{Template: TokenCard("5/5 white Giant Warrior"), N: 1}.Apply(ctx)
			}),
		),
	})
}
