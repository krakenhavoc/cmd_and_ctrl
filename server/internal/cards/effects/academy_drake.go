package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Academy Drake — Creature — Drake {2}{U}, 2/2 (Dominaria, #5530):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Flying
//	 If this creature was kicked, it enters with two +1/+1 counters on it."
//
// Shalai's Acolyte's kicker: a kicked Drake enters with two +1/+1
// counters; a single kicker is paid at most once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "65f3ebeb-1ca5-4923-b681-032579991c8a",
		Name:                       "Academy Drake",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		OptionalCosts:              []game.AdditionalCost{Kicker("{4}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
	})
}
