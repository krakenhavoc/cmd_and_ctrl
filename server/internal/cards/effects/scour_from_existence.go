package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scour from Existence — Instant {7} (Battle for Zendikar, #5547):
//
//	"Exile target permanent."
//
// Any permanent may be the target, lands included, checked again at
// resolution (CR 608.2b); exile, so indestructible doesn't save it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "241a50c5-f65f-4847-89c7-5c0ef6025dc1",
		Name:         "Scour from Existence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
