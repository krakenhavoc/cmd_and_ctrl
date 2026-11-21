package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grixis Panorama — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 Island, Swamp, or Mountain card, put it onto the battlefield tapped,
//	 then shuffle."
//
// A row of the Panorama cycle (b43PanoramaFetch, see Naya Panorama).
//
// No simplification.
func init() {
	effects.Register(effects.B43PanoramaFetch(
		"743f4488-fef1-4f4d-b745-d2de92423e00", "Grixis Panorama",
		"Island", "Swamp", "Mountain"))
}
