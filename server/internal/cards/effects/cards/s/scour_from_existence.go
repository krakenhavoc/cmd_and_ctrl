package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scour from Existence — Instant {7} (Battle for Zendikar, #5547):
//
//	"Exile target permanent."
//
// Any permanent may be the target, lands included, checked again at
// resolution (CR 608.2b); exile, so indestructible doesn't save it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "241a50c5-f65f-4847-89c7-5c0ef6025dc1",
		Name:         "Scour from Existence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
