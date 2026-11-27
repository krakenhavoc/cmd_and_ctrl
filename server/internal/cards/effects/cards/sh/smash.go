package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smash — Instant {2}{R} (Ravnica: City of Guilds):
//
//	"Destroy target artifact.
//	 Draw a card."
//
// Shatter's sentence (CR 701.8a) on the shared destroy-the-target body, then the
// draw, in printed order. With its one target illegal on resolution the whole
// spell does nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1602f8c7-fe00-4bc9-b4fc-b26b7223ac75",
		Name:         "Smash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
