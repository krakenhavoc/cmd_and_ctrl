package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8a39809e-219e-4f59-a233-ea28d05c30ef",
		Name:         "Quandrix Pledgemage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Quandrix Pledgemage — put a +1/+1 counter on it (magecraft)",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
