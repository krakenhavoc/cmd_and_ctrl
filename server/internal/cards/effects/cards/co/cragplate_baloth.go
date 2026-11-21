package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cragplate Baloth — Creature — Beast {5}{G}{G}, 6/6 (Zendikar Rising,
// #5508):
//
//	"Kicker {2}{G}
//	 This spell can't be countered.
//	 Hexproof, haste
//	 If this creature was kicked, it enters with four +1/+1 counters on
//	 it."
//
// Gigantic Big Bear's body (can't be countered, hexproof, haste) with
// Vodalian Serpent's kicked entry: the counters are a CR 614.1c entry
// clause read off the cast (CountersPerKick), so a kicked Baloth enters
// with four and an unkicked one, or one put onto the battlefield without
// being cast, with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3f599e15-cbc5-4170-b6b4-efd16aa222ac",
		Name:                       "Cragplate Baloth",
		Completeness:               effects.CompletenessFull,
		CantBeCountered:            true,
		PrintedKeywords:            []string{"hexproof", "haste"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 4)},
	})
}
