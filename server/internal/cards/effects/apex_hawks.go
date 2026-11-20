package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apex Hawks — Creature — Bird {2}{W}, 2/2 (Worldwake, #5582):
//
//	"Multikicker {1}{W} (You may pay an additional {1}{W} any number of times as
//	 you cast this spell.)
//	 Flying
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
		OracleID:                   "c025d558-aae8-4146-923b-a98035025a42",
		Name:                       "Apex Hawks",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		OptionalCosts:              []game.AdditionalCost{Multikicker("{1}{W}", 20)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 1)},
	})
}
