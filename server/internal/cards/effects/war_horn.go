package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Horn — Artifact {3} (Magic Origins, #5549):
//
//	"Attacking creatures you control get +1/+0."
//
// Goblin Oriflamme's static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99bc739e-c745-47d8-8401-2f34b6becd85",
		Name:         "War Horn",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{attackingCreaturesYouControlGetPower(1)},
	})
}
