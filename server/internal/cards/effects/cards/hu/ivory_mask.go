package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivory Mask — Enchantment {2}{W}{W} (Eighth Edition, #5607):
//
//	"You have shroud. (You can't be the target of spells or abilities.)"
//
// Shroud on its controller (CR 702.18a), the player-keyword Solitary
// Confinement uses; it stops the controller's own spells too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "028bc726-0e86-4c53-9648-625153ae4b01",
		Name:           "Ivory Mask",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordShroud},
	})
}
