package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inner-Chamber Guard — Creature — Human Samurai {1}{W}, 0/2 (Saviors of Kamigawa, #5601):
//
//	"Bushido 2"
//
// The printed keywords are engine data; Bushido 2 is the shared keyword
// trigger (CR 702.45a).
//
// No simplification.
func init() {
	effects.ChkB1Bushido("ea5e8920-1343-4335-b516-6db43c3bfeb0", "Inner-Chamber Guard", 2)
}
