package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final Reward — Instant {4}{B}
// (Amonkhet, #5535):
//
//	"Exile target creature."
//
// Wander Off's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e654242f-c7c5-4713-bbd0-26d41de8e2e7",
		Name:         "Final Reward",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
