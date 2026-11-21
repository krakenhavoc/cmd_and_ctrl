package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secret Passage — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.BnsB21ClueTapland("c6a46fc2-fc8f-4dcd-bca7-682abfbf303d", "Secret Passage", "U", "B"))
}
