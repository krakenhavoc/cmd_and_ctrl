package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quandrix Pledgemage — Creature — Merfolk Druid {1}{G/U}{G/U}, 2/2
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, put
//	 a +1/+1 counter on this creature."
//
// The shared Magecraft trigger with Ajani's Pridemate's counter on the
// source; a Pledgemage that has left the battlefield gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a39809e-219e-4f59-a233-ea28d05c30ef",
		Name:         "Quandrix Pledgemage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Quandrix Pledgemage — put a +1/+1 counter on it (magecraft)",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
