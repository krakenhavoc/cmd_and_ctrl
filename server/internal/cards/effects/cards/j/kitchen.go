package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitchen — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.BnsB21ClueTapland("47f71408-5509-4651-9121-fd0867adae00", "Kitchen", "G", "U"))
}
