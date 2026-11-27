package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Body of Jukai — Creature — Spirit {7}{G}{G}, 8/5 (Betrayers of Kamigawa, S82 OLD b596):
//
//	"Trample
//	 Soulshift 8"
//
// Soulshift 8 is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("90d33e75-17bd-4fc9-82f5-eff254135df5", "Body of Jukai", 8, "trample")
}
