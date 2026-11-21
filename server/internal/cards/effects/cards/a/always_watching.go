package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Always Watching — Enchantment {1}{W}{W}:
//
//	"Nontoken creatures you control get +1/+1 and have vigilance."
//
// A layer 7c anthem and a layer 6 keyword grant over the same set.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2153115-bc12-458d-baeb-538ec61e1242",
		Name:         "Always Watching",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CmdrB1Labeled(effects.B16Anthem(effects.CmdrB1NontokenCreaturesYouControl, 1, 1), "Nontoken creatures you control get +1/+1."),
			effects.CmdrB1Labeled(effects.B16GrantKeywords(effects.CmdrB1NontokenCreaturesYouControl, "vigilance"), "Nontoken creatures you control have vigilance."),
		},
	})
}
