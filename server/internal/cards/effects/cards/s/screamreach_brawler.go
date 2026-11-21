package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screamreach Brawler — Creature — Orc Berserker {2}{R}, 2/3 (Dragons of Tarkir,
// #5551):
//
//	"Dash {1}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
//
// Dash is the shared alternative cost (Ragavan's, CR 702.109): cast for
// its dash cost it gains haste and returns to its owner's hand at the
// beginning of the next end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ffc84ff4-22eb-4652-9723-2f40d637a427",
		Name:             "Screamreach Brawler",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{1}{R}")},
	})
}
