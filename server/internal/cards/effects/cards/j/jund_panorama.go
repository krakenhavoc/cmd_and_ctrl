package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jund Panorama — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 Swamp, Mountain, or Forest card, put it onto the battlefield
//	 tapped, then shuffle."
//
// The Alara Panorama cycle's Jund row: Naya Panorama's table entry
// (b43PanoramaFetch) with the three Jund basic types, in the printed
// order. The sacrifice is part of the cost; the search happens on
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.B43PanoramaFetch(
		"f39f33ac-074d-442d-ae4c-1d694ee315f3", "Jund Panorama",
		"Swamp", "Mountain", "Forest"))
}
