package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tinder Farm — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {T}, Sacrifice this land: Add {R}{W}."
//
// One of Invasion's five sacrifice lands, built by OldB663SacrificeLand: the
// tapped entry and the tap for one color of TappedSacrificeLand, with a
// one-shot tap-and-sacrifice for the two colors it names.
//
// No simplification.
func init() {
	effects.Register(effects.OldB663SacrificeLand("af0ad159-8264-4526-9a95-eddd32c0a13f", "Tinder Farm", "G", "RW"))
}
