package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paragon of Fierce Defiance — Creature — Human Warrior {3}{R}, 2/2
// (Magic 2015, #5556):
//
//	"Other red creatures you control get +1/+1.
//	 {R}, {T}: Another target red creature you control gains haste until
//	 end of turn."
//
// The M15 Paragon cycle's shared body (m15B1Paragon).
//
// No simplification.
func init() {
	effects.Register(effects.M15B1Paragon("0f94d1a7-351f-4d37-8dab-1ebf2e8ee213", "Paragon of Fierce Defiance", "{R}", "R", "red", "haste"))
}
