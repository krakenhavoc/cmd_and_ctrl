package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4dacd89b-5983-4905-9f6c-43aa6a12149d",
		Name:         "Feudkiller's Verdict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.GainLife{Amount: 10},
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.MorMoreLifeThanAnOpponent(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("5/5 white Giant Warrior"), N: 1}.Apply(ctx)
			}),
		),
	})
}
