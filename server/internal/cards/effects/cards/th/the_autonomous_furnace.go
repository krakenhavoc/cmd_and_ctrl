package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Autonomous Furnace — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {1}{R}, {T}, Sacrifice this land: Draw a card."
//
// The tapped entry is a real CR 614 self-replacement, the mana a mana
// ability, and the draw an activated ability whose cost taps and
// sacrifices the land (CR 602.2).
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSphereDrawLand("e2dd05d5-312e-47f1-873e-c0741ee6ef4a", "The Autonomous Furnace", "R"))
}
