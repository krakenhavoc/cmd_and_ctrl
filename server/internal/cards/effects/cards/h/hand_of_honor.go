package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hand of Honor — Creature — Human Samurai {W}{W}, 2/2 (Saviors of Kamigawa, #5601):
//
//	"Protection from black
//	 Bushido 1"
//
// The printed keywords are engine data; Bushido 1 is the shared keyword
// trigger (CR 702.45a).
//
// No simplification.
func init() {
	effects.ChkB1Bushido("feb65f2e-6f71-4f29-bf31-b210c85f3ab3", "Hand of Honor", 1, "protection from black")
}
