package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geothermal Crevice — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {T}, Sacrifice this land: Add {B}{G}."
//
// One of Invasion's five sacrifice lands, built by OldB663SacrificeLand: the
// tapped entry and the tap for one color of TappedSacrificeLand, with a
// one-shot tap-and-sacrifice for the two colors it names.
//
// No simplification.
func init() {
	effects.Register(effects.OldB663SacrificeLand("83f9d84a-a011-4fa9-a3a4-51eb39159c54", "Geothermal Crevice", "R", "BG"))
}
