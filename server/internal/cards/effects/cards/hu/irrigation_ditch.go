package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irrigation Ditch — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {T}, Sacrifice this land: Add {G}{U}."
//
// One of Invasion's five sacrifice lands, built by OldB663SacrificeLand: the
// tapped entry and the tap for one color of TappedSacrificeLand, with a
// one-shot tap-and-sacrifice for the two colors it names.
//
// No simplification.
func init() {
	effects.Register(effects.OldB663SacrificeLand("e0c3c87b-83e3-4cff-b05f-78a192e2bba4", "Irrigation Ditch", "W", "GU"))
}
