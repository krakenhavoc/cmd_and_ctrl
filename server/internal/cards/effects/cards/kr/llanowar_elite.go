package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Elite — Creature — Elf {G}, 1/1 (Invasion, S82 OLD #5616):
//
//	"Kicker {8} (You may pay an additional {8} as you cast this spell.)
//	 Trample
//	 If this creature was kicked, it enters with five +1/+1 counters on it."
//
// Baloth Gorger's kicker and entry counters (CR 614.1c); trample is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "9e112393-962b-461e-a605-9f8d6a1ebaa4",
		Name:                       "Llanowar Elite",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"trample"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{8}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 5)},
	})
}
