package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Water Servant — Creature — Elemental {2}{U}{U}, 3/4 (Magic 2011, #5579):
//
//	"{U}: This creature gets +1/-1 until end of turn.
//	 {U}: This creature gets -1/+1 until end of turn."
//
// Thornling's two trades through the shared thisGetsRow, each a layer 7c
// change that ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8c32c80-9693-4ae4-a387-21ec8a3cdd40",
		Name:         "Water Servant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ThisGetsRow("{U}", 1, -1, "Water Servant"),
			effects.ThisGetsRow("{U}", -1, 1, "Water Servant"),
		},
	})
}
