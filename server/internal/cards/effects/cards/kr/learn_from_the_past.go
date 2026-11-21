package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Learn from the Past — Instant {3}{U} (Tales of Middle-earth Commander):
//
//	"Target player shuffles their graveyard into their library.
//	 Draw a card."
//
// Two steps: the target player's graveyard is shuffled into their library
// (every card moves, then the library is shuffled), then you draw a card; the
// draw is yours, not the target's, and happens even if the target is gone by
// resolution (CR 608.2b: the shuffle does nothing, the draw still does).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f325fd7a-dcae-4892-adf5-de9893a23ac5",
		Name:         "Learn from the Past",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, t.ID); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
