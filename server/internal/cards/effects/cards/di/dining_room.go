package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dining Room — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.BnsB21ClueTapland("0880461e-8943-443b-90e7-ff84eef46550", "Dining Room", "R", "G"))
}
