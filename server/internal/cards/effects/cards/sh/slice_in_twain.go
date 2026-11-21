package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slice in Twain — Instant {2}{G}{G} (Scars of Mirrodin, #5577):
//
//	"Destroy target artifact or enchantment.
//	 Draw a card."
//
// Break Asunder's destroy, then a draw. If the target is gone by
// resolution the whole spell does nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a96d0065-b8c1-4843-8752-dce60bf83545",
		Name:         "Slice in Twain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
