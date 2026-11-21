package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Empty Graves — Creature — Spirit {3}{B}, 4/1 (Saviors of Kamigawa, #5601):
//
//	"Soulshift 3"
//
// Soulshift 3 is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("bff92baa-fc0e-4da4-8bd5-e57e4963d642", "Kami of Empty Graves", 3)
}
