package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timberland Ruins — Land (Odyssey, #5613):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// One of Odyssey's five sacrifice lands, built by odySacrificeLand:
// an enters-tapped replacement and two mana abilities, the second
// paying with the land itself.
//
// No simplification.
func init() {
	effects.Register(effects.OdySacrificeLand("29c7f059-2eeb-40f9-8f50-03eba2d0d5e0", "Timberland Ruins", "{G}"))
}
