package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Untamed Kavu — Creature — Kavu {1}{G}, 2/2 (Dominaria, #5530):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 Vigilance, trample
//	 If this creature was kicked, it enters with three +1/+1 counters on it."
//
// Baloth Gorger's kicker: a kicked Kavu enters with three +1/+1 counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "ce585b28-f22f-41ba-8858-aebc0164922d",
		Name:                       "Untamed Kavu",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"vigilance", "trample"},
		OptionalCosts:              []game.AdditionalCost{Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 3)},
	})
}
