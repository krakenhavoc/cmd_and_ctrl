package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Always Watching — Enchantment {1}{W}{W}:
//
//	"Nontoken creatures you control get +1/+1 and have vigilance."
//
// A layer 7c anthem and a layer 6 keyword grant over the same set.
func init() {
	Register(Spec{
		OracleID:     "c2153115-bc12-458d-baeb-538ec61e1242",
		Name:         "Always Watching",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			cmdrB1Labeled(b16Anthem(cmdrB1NontokenCreaturesYouControl, 1, 1), "Nontoken creatures you control get +1/+1."),
			cmdrB1Labeled(b16GrantKeywords(cmdrB1NontokenCreaturesYouControl, "vigilance"), "Nontoken creatures you control have vigilance."),
		},
	})
}
