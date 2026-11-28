package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Updraft — Instant {1}{U} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Target creature gains flying until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// A layer 6 keyword grant that ends at cleanup (CR 514.2), then the delayed draw
// at the next upkeep (CR 603.7a, next_upkeep_draw.go). With its one target gone
// the spell does not resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff0d5cbe-99d3-4d3b-9a5b-80223fc18869",
		Name:         "Updraft",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.EbTargetCreatureGainsUntilEOT("flying", "Updraft — flying until end of turn")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Updraft — draw a card").Apply(ctx)
		},
	})
}
