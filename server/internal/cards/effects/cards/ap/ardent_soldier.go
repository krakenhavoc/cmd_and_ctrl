package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardent Soldier — Creature — Human Soldier {1}{W}, 1/2 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Vigilance
//	 If this creature was kicked, it enters with a +1/+1 counter on it."
//
// Baloth Gorger's kicker and entry counter (CR 614.1c); vigilance is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "a56d301b-8f4c-4f12-8ee0-34c6b66faee0",
		Name:                       "Ardent Soldier",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"vigilance"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
	})
}
