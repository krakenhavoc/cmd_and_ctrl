package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Hunter Maze — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {1}{G}, {T}, Sacrifice this land: Draw a card."
//
// The tapped entry is a real CR 614 self-replacement, the mana a mana
// ability, and the draw an activated ability whose cost taps and
// sacrifices the land (CR 602.2).
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSphereDrawLand("b0ea4975-a944-4585-8711-60f203cffa4a", "The Hunter Maze", "G"))
}
