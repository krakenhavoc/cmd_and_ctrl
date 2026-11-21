package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enclave Elite — Creature — Merfolk Soldier {2}{U}, 2/2 (Worldwake, #5582):
//
//	"Multikicker {1}{U} (You may pay an additional {1}{U} any number of times as
//	 you cast this spell.)
//	 Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)
//	 This creature enters with a +1/+1 counter on it for each time it
//	 was kicked."
//
// Everflowing Chalice's two primitives: Multikicker for the repeatable
// cost and CountersPerKick for the entry, one +1/+1 counter per kick
// (CR 702.33d). The cap of 20 is the engine's, not the card's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "27c84b2e-fc66-45a6-a5ad-d3ef204d5077",
		Name:                       "Enclave Elite",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"islandwalk"},
		OptionalCosts:              []game.AdditionalCost{effects.Multikicker("{1}{U}", 20)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
