package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crawling Filth — Creature — Spirit {5}{B}, 2/2 (Betrayers of Kamigawa, S82 OLD b596):
//
//	"Fear
//	 Soulshift 5"
//
// Soulshift 5 is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("d76522e0-324f-4ec9-b36e-0fe7448aae69", "Crawling Filth", 5, "fear")
}
