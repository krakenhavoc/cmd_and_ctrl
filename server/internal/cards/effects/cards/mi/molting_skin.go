package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molting Skin — {2}{G} Enchantment:
//
//	"Return this enchantment to its owner's hand: Regenerate target
//	 creature."
//
// #2028: Broken Fall's sentence, through the same row
// (returnThisRegenerateRow).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8262dbb3-7abb-4de8-9a1d-edee71906a10",
		Name:         "Molting Skin",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ReturnThisRegenerateRow()},
	})
}
