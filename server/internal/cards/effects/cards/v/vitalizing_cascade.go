package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vitalizing Cascade — Instant {X}{G}{W} (Mirage, S82 OLD batch 809, #5633):
//
//	"You gain X plus 3 life."
//
// X is the value the cost engine charged (CR 107.3); the gain is X + 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaf176ff-646f-4182-a546-fe648498587d",
		Name:         "Vitalizing Cascade",
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: ctx.X() + 3}.Apply(ctx)
		},
	})
}
