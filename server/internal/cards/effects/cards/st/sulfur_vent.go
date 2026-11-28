package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sulfur Vent — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {T}, Sacrifice this land: Add {U}{R}."
//
// One of Invasion's five sacrifice lands, built by OldB663SacrificeLand: the
// tapped entry and the tap for one color of TappedSacrificeLand, with a
// one-shot tap-and-sacrifice for the two colors it names.
//
// No simplification.
func init() {
	effects.Register(effects.OldB663SacrificeLand("bc2906cd-9e0c-4aa8-b656-fb1045d810ac", "Sulfur Vent", "B", "UR"))
}
