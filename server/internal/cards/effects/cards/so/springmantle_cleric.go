package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Springmantle Cleric — Creature — Elf Cleric {4}{G}, 2/3 (Zendikar Rising,
// S82 OLD #5508):
//
//	"This creature enters with a +1/+1 counter on it for each color of mana
//	 spent to cast it."
//
// A CR 614.1c entry clause read off the payment of the resolving spell
// (Wildgrowth Archaic's), so it composes with every other counter effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "83865eee-709a-4a34-94f1-9b818f31323f",
		Name:                "Springmantle Cleric",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerColorSpent(game.CounterPlusOne),
		},
	})
}
