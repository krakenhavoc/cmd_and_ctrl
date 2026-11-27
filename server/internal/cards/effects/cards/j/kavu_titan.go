package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Titan — Creature — Kavu {1}{G}, 2/2 (Invasion, S82 OLD #5616):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
//	 If this creature was kicked, it enters with three +1/+1 counters on it
//	 and with trample."
//
// Benalish Lancer's shape: the kicked entry counters (CR 614.1c) and a layer 6
// keyword grant keyed on the permanent's own kicker record (CR 400.7d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "df166d0e-f944-404d-83b9-8cf53e220774",
		Name:                       "Kavu Titan",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 3)},
		Static:                     []game.StaticAbility{effects.OldB650KickedKeywordGrant("trample")},
	})
}
