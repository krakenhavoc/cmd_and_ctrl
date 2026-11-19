package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baloth Gorger — Creature — Beast {2}{G}{G}, 4/4 (Foundations Jumpstart,
// #5929):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 If this creature was kicked, it enters with three +1/+1 counters on
//	 it."
//
// Gnarlid Colony's kicker and counters read off the cast: a single kicker
// is paid at most once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "7f9bc0c9-0333-4c02-bc14-e6affe848b5d",
		Name:                       "Baloth Gorger",
		Completeness:               CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{Kicker("{4}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 3)},
	})
}
