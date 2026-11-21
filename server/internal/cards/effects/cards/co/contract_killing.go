package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contract Killing — Sorcery {3}{B}{B} (Foundations Jumpstart):
//
//	"Destroy target creature. Create two Treasure tokens."
//
// Bake into a Pie's shape with two Treasures: the tokens are made even
// when the creature survives the destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7a0e473-77c1-4a85-b6d7-b0dfdadb696f",
		Name:         "Contract Killing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.CreateToken{Template: effects.TreasureToken(), N: 2},
		),
	})
}
