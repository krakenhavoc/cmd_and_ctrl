package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hand of Cruelty — Creature — Human Samurai {B}{B}, 2/2 (Saviors of Kamigawa, #5601):
//
//	"Protection from white
//	 Bushido 1"
//
// The printed keywords are engine data; Bushido 1 is the shared keyword
// trigger (CR 702.45a).
//
// No simplification.
func init() {
	effects.ChkB1Bushido("ddbd4d8c-9763-43d7-882f-6edf8efffb0a", "Hand of Cruelty", 1, "protection from white")
}
