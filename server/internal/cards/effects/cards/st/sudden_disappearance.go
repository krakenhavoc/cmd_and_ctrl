package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Disappearance — Sorcery {5}{W}:
//
//	"Exile all nonland permanents target player controls. Return the exiled
//	 cards to the battlefield under their owner's control at the beginning of
//	 the next end step."
//
// One exile of the whole set, and the delayed return over the cards that reached
// exile (a token is gone for good, CR 111.7). They come back under their owners'
// control, as new objects (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbe28518-d795-4bc5-afa6-63c696c86fca",
		Name:         "Sudden Disappearance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, Partial: true}},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := effects.TargetedPlayerOf(ctx)
			if p == uuid.Nil {
				return nil
			}
			return effects.ExileCardsThenScheduleReturn(ctx, effects.NonlandPermanentsControlledBy(ctx.Game, p), "Sudden Disappearance — return the exiled cards")
		},
	})
}
