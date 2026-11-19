package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contract Killing — Sorcery {3}{B}{B} (Foundations Jumpstart):
//
//	"Destroy target creature. Create two Treasure tokens."
//
// Bake into a Pie's shape with two Treasures: the tokens are made even
// when the creature survives the destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7a0e473-77c1-4a85-b6d7-b0dfdadb696f",
		Name:         "Contract Killing",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyTheTargetPermanent(ctx.Item, ctx) }),
			CreateToken{Template: TreasureToken(), N: 2},
		),
	})
}
