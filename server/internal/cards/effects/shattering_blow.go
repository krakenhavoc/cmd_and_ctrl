package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shattering Blow — Instant {1}{R/W} (Gatecrash, #5565):
//
//	"Exile target artifact."
//
// Erase's spell with an artifact target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc9374df-134d-4d32-842e-991e2bfb2dc1",
		Name:         "Shattering Blow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
