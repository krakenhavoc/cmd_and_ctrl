package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Fair Basilica — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {1}{W}, {T}, Sacrifice this land: Draw a card."
//
// The tapped entry is a real CR 614 self-replacement, the mana a mana
// ability, and the draw an activated ability whose cost taps and
// sacrifices the land (CR 602.2).
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSphereDrawLand("8618f8e8-fd86-47e3-905d-e6624b599b9b", "The Fair Basilica", "W"))
}
