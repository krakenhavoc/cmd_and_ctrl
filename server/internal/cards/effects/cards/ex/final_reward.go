package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Reward — Instant {4}{B}
// (Amonkhet, #5535):
//
//	"Exile target creature."
//
// Wander Off's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e654242f-c7c5-4713-bbd0-26d41de8e2e7",
		Name:         "Final Reward",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
