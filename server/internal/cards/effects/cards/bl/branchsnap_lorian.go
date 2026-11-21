package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Branchsnap Lorian — Creature — Beast {1}{G}{G}, 4/1 (Legions, #5609):
//
//	"Trample
//	 Morph {G}"
//
// Battering Craghorn's shape: face down it is a 2/2 with no trample
// (CR 708.2a); turned face up for {G} it is the 4/1 trampler, the same
// object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e09b24a0-675b-4c64-ac44-74ddf109598a",
		Name:             "Branchsnap Lorian",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{G}")},
	})
}
