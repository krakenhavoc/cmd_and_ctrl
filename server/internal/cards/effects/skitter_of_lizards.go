package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skitter of Lizards — Creature — Lizard {R}, 1/1 (Worldwake, #5582):
//
//	"Multikicker {1}{R} (You may pay an additional {1}{R} any number of times as
//	 you cast this spell.)
//	 Haste
//	 This creature enters with a +1/+1 counter on it for each time it
//	 was kicked."
//
// Everflowing Chalice's two primitives: Multikicker for the repeatable
// cost and CountersPerKick for the entry, one +1/+1 counter per kick
// (CR 702.33d). The cap of 20 is the engine's, not the card's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "14646c85-12ad-4b99-a2af-e4a790e034ef",
		Name:                       "Skitter of Lizards",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"haste"},
		OptionalCosts:              []game.AdditionalCost{Multikicker("{1}{R}", 20)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 1)},
	})
}
