package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gone Missing — Sorcery {4}{U} (Shadows over Innistrad, #5543):
//
//	"Put target permanent on top of its owner's library.
//	 Investigate."
//
// Banishment Decree's tuck with any permanent, then the Clue: the permanent
// goes on top of its owner's library, not its controller's, and only a target
// that is illegal as the spell resolves stops the whole spell (CR 608.2b), so
// the Clue is made even if the permanent could not be moved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a31a66ee-5ddf-4d70-9ae9-725fc7327706",
		Name:         "Gone Missing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.PutTargetOnTopOfOwnersLibrary(ctx.Item, ctx) }),
			effects.Investigate{N: 1},
		),
	})
}
