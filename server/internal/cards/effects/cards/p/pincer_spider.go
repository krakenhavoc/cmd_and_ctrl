package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pincer Spider — Creature — Spider {2}{G}, 2/3 (Invasion, S82 OLD #5616):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 Reach (This creature can block creatures with flying.)
//	 If this creature was kicked, it enters with a +1/+1 counter on it."
//
// Baloth Gorger's kicker and entry counter (CR 614.1c); reach is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "fbe25c32-db20-464c-a0ed-d90e75b111a9",
		Name:                       "Pincer Spider",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"reach"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
