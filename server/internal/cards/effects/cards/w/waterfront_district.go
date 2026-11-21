package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterfront District — Land (Streets of New Capenna, #5491):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {2}{U}{B}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's five sacrifice-lands, registered through
// sncCyclingTapland (s82_snc_b03_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.SncCyclingTapland("18b0b583-d17b-4c67-bc67-c4d9aad155c7", "Waterfront District", "U", "B"))
}
