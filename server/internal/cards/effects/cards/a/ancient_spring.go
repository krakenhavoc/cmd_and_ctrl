package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Spring — Land (Invasion, S82 older sets batch 663, #5616):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {T}, Sacrifice this land: Add {W}{B}."
//
// One of Invasion's five sacrifice lands, built by OldB663SacrificeLand: the
// tapped entry and the tap for one color of TappedSacrificeLand, with a
// one-shot tap-and-sacrifice for the two colors it names.
//
// No simplification.
func init() {
	effects.Register(effects.OldB663SacrificeLand("2160849a-184f-4a49-8931-fd021e16f7cb", "Ancient Spring", "U", "WB"))
}
