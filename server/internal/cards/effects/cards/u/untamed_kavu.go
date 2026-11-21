package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:                   "ce585b28-f22f-41ba-8858-aebc0164922d",
		Name:                       "Untamed Kavu",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"vigilance", "trample"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 3)},
	})
}
