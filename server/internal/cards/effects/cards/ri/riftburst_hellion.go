package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riftburst Hellion — Creature — Hellion {5}{R}{G}
// (Murders at Karlov Manor, #4469):
//
//	"Reach
//	 Disguise {4}{R/G}{R/G} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Nothing beyond the keyword and the disguise declaration (CR 702.168a);
// the hybrid face-up cost is paid like any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e121f9cb-ed25-4be1-9822-b68cce2f2623",
		Name:             "Riftburst Hellion",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{4}{R/G}{R/G}")},
	})
}
