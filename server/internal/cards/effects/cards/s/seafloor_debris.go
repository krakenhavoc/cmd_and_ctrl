package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seafloor Debris — Land (Odyssey, #5613):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// One of Odyssey's five sacrifice lands, built by odySacrificeLand:
// an enters-tapped replacement and two mana abilities, the second
// paying with the land itself.
//
// No simplification.
func init() {
	effects.Register(effects.OdySacrificeLand("eef64810-4187-40c1-8e59-1cb8773ebea3", "Seafloor Debris", "{U}"))
}
