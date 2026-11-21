package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Library — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.BnsB21ClueTapland("33a7d29a-ebc0-4606-91b6-7381e2a16016", "Library", "U", "R"))
}
