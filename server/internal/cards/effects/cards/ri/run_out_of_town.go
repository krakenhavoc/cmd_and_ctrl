package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Run Out of Town — Instant {3}{U} (Streets of New Capenna, #5491):
//
//	"The owner of target nonland permanent puts it on their choice of the
//	 top or bottom of their library."
//
// Desynchronize without the scry: the owner picks the lane.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96189b18-8cab-4709-b3b6-c4f2003829eb",
		Name:         "Run Out of Town",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Run Out of Town — put it on your choice of the top or bottom of your library")
		},
	})
}
