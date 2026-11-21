package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattering Blow — Instant {1}{R/W} (Gatecrash, #5565):
//
//	"Exile target artifact."
//
// Erase's spell with an artifact target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc9374df-134d-4d32-842e-991e2bfb2dc1",
		Name:         "Shattering Blow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
