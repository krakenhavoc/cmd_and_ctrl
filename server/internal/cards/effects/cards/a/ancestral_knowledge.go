package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancestral Knowledge — {1}{U} Enchantment (Weatherlight, S82 OLD batch 772,
// #5629):
//
//	"Cumulative upkeep {1}
//	 When this enchantment enters, look at the top ten cards of your library,
//	 then exile any number of them and put the rest back on top of your library
//	 in any order.
//	 When this enchantment leaves the battlefield, shuffle your library."
//
// Two triggers beside the cumulative upkeep. The enter trigger is a look (only its
// controller learns the cards, CR 701.20) at the top ten (fewer when the library
// is short), a pick of any number of them from none to all, the picked cards
// exiled and the rest put back as an ordering prompt. The leave trigger covers
// every exit, the sacrifice for an unpaid upkeep included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2786f210-ec64-46ea-9a7f-b03759fb1dae",
		Name:         "Ancestral Knowledge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Ancestral Knowledge — cumulative upkeep {1}", "{1}"),
			effects.WhenThisEnters("Ancestral Knowledge — look at the top ten cards of your library, then exile any number of them and put the rest back on top in any order",
				effects.S82OldB772LookAtTopTenExileAnyNumber),
			effects.WhenThisLeaves("Ancestral Knowledge — shuffle your library", effects.Do(effects.ShuffleLibrary{})),
		},
	})
}
