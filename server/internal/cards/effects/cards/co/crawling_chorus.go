package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crawling Chorus — Creature — Phyrexian Horror {W}, 1/1: "Toxic 1. When this
// creature dies, create a 1/1 colorless Phyrexian Mite artifact creature token
// with toxic 1 and 'This token can't block.'" No simplification.
func init() {
	effects.Register(effects.S82OneB04ToxicDiesMite("99ba0d4f-b2bf-48ea-95d6-b6b85ee077ab", "Crawling Chorus"))
}
