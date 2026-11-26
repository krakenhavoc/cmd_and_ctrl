package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistcutter Hydra — Creature — Hydra {X}{G}, 0/0 (Theros):
//
//	"This spell can't be countered.
//	 Haste, protection from blue
//	 This creature enters with X +1/+1 counters on it."
//
// "Can't be countered" is Spec.CantBeCountered (Skylasher's), the X counters
// are the printed CR 614.1c entry clause (XCounters, Benevolent Hydra's), and
// haste and protection from blue are printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "30b5b764-478a-46fa-8f95-a1f515925a2d",
		Name:                       "Mistcutter Hydra",
		Completeness:               effects.CompletenessFull,
		CantBeCountered:            true,
		XMatters:                   true,
		PrintedKeywords:            []string{"haste", "protection from blue"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
	})
}
