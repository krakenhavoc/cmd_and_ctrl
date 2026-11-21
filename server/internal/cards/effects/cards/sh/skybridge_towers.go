package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skybridge Towers — Land (Streets of New Capenna, #5491):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {2}{W}{U}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's five sacrifice-lands, registered through
// sncCyclingTapland (s82_snc_b03_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.SncCyclingTapland("46466819-3c8b-4f7e-82c2-93cb0118a6e6", "Skybridge Towers", "W", "U"))
}
