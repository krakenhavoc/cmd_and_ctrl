package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Familiar Ground — Enchantment {2}{G} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Each creature you control can't be blocked by more than one creature."
//
// Bristling Boar's block rule over every creature its controller controls (CR 509.1b), read
// live at each declaration, so a creature that came under their control later is bound too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b92c4dc-2605-42f4-a374-c3be4de3279e",
		Name:         "Familiar Ground",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.ControlledBySourceController(), 1),
		},
	})
}
