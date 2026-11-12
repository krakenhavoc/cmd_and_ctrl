package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dust to Dust — Sorcery {1}{W}{W} (The Dark, #5638):
//
//	"Exile two target artifacts."
//
// Ashes to Ashes' exactly-two clause on artifacts, with no damage
// tail: each target still legal as it resolves is exiled, so one
// that left in response is skipped and the other is still exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b17eb586-0272-4d02-be75-de06ea044b6f",
		Name:         "Dust to Dust",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: exiles permanents
		Targets:      TargetPermanent("two target artifacts", Artifact()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileEveryLegalCardTarget(ctx.Game, item)
		},
	})
}
