package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slay — Instant {2}{B} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"Destroy target green creature. It can't be regenerated.
//	 Draw a card."
//
// Befoul's destroy without regeneration, then the draw as a keyed tail behind
// it (Leave in the Dust's shape). If the target is illegal as this resolves
// the whole spell does nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64418e1f-4fb1-4029-a31e-135b2d2c0ab9",
		Name:         "Slay",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetCreature("target green creature", effects.OfColor("G")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := effects.DestroyTheTargetPermanentNoRegen(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
