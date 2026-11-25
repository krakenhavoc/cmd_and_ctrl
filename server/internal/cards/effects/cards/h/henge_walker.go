package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Henge Walker — Artifact Creature — Golem {3}, 2/2 (Throne of Eldraine,
// #5516):
//
//	"Adamant — If at least three mana of the same color was spent to cast
//	 this spell, this creature enters with a +1/+1 counter on it."
//
// The adamant entry counter with any one colour: at least three mana of a
// single colour among what paid for it (colourless is not a colour,
// CR 105.1), read off the spell's payment as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "d4c13349-2faa-46ec-bd8e-80a4e59ef407",
		Name:                       "Henge Walker",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfThreeOfOneColorSpent(game.CounterPlusOne)},
	})
}
