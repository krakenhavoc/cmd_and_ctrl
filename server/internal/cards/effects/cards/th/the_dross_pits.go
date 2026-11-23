package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Dross Pits — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {1}{B}, {T}, Sacrifice this land: Draw a card."
//
// The tapped entry is a real CR 614 self-replacement, the mana a mana
// ability, and the draw an activated ability whose cost taps and
// sacrifices the land (CR 602.2).
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSphereDrawLand("8110fe69-c66c-4e2c-86ee-dcc8dc9a13d1", "The Dross Pits", "B"))
}
