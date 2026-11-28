package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stun — Instant {1}{R} (Invasion, S82 #5616):
//
//	"Target creature can't block this turn.
//	 Draw a card."
//
// Blindblast's "can't block this turn" on its own, with the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3491972-44c5-4962-a680-42b79357a189",
		Name:         "Stun",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)).Targets},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := ctx.ClauseTarget(0); !ok {
				return nil
			}
			if err := effects.TargetCreatureCantBlockThisTurn("Stun — that creature can't block this turn")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
