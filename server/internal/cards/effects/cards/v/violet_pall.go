package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Violet Pall — Kindred Instant — Faerie {4}{B} (Morningtide, #5592):
//
//	"Destroy target nonblack creature. Create a 1/1 black Faerie Rogue
//	 creature token with flying."
//
// Corpsehatch's shape: the destroy, then the token. With its only target
// gone the spell does nothing at all (CR 608.2b), so no token either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf6ea336-6a4d-4653-acdd-f09af0b9f0e3",
		Name:         "Violet Pall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, ctx.Item)
		}), effects.CreateToken{Template: effects.FaerieRogueToken(), N: 1}),
	})
}
