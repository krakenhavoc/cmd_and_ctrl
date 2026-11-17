package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shalai's Acolyte — Creature — Angel {4}{W}, 3/4 (Dominaria United,
// #5487):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this
//	 spell.)
//	 Flying
//	 If this creature was kicked, it enters with two +1/+1 counters on
//	 it."
//
// Vodalian Serpent's CR 614.1c entry clause read off the cast
// (CountersPerKick); a Shalai's Acolyte put onto the battlefield any other
// way was not kicked and enters with none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "53ec394a-5aaa-4d83-82ec-f00e99902c0d",
		Name:                       "Shalai's Acolyte",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		OptionalCosts:              []game.AdditionalCost{Kicker("{1}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
	})
}
