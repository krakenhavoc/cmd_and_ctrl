package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Araba Mothrider — Creature — Human Samurai {1}{W}, 1/1 (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 Bushido 1"
//
// The printed keywords are engine data; Bushido 1 is the shared keyword
// trigger (CR 702.45a).
//
// No simplification.
func init() {
	effects.ChkB1Bushido("d6b54f58-7775-4180-b2e5-1b57da65a409", "Araba Mothrider", 1, "flying")
}
