package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enervate — Instant {1}{U} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Tap target artifact, creature, or land.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Ring of Gix's target and tap, then the delayed draw at the next upkeep
// (CR 603.7a, next_upkeep_draw.go). With its one target gone the spell does not
// resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f763d95e-87af-473f-9b1f-cf16221e5792",
		Name:         "Enervate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.TapEach(ctx, effects.HoldTargetIDs(ctx)); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Enervate — draw a card").Apply(ctx)
		},
	})
}
