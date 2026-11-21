package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paragon of Gathering Mists — Creature — Human Wizard {3}{U}, 2/2 (Magic
// 2015, #5556):
//
//	"Other blue creatures you control get +1/+1.
//	 {U}, {T}: Another target blue creature you control gains flying until
//	 end of turn."
//
// The M15 Paragon cycle's shared body (m15B1Paragon).
//
// No simplification.
func init() {
	effects.Register(effects.M15B1Paragon("38212d24-40fd-404f-a8bb-24ed0d023fe8", "Paragon of Gathering Mists", "{U}", "U", "blue", "flying"))
}
