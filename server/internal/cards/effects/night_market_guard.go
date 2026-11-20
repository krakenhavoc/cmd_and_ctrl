package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night Market Guard — Artifact Creature — Construct {3}, 3/1 (Aether Revolt,
// #5536):
//
//	"This creature can block an additional creature each combat."
//
// Foriysian Brigade's block capacity of two (CR 509.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a7a27fe-60f0-4a21-a1c1-5fef8a528391",
		Name:         "Night Market Guard",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
	})
}
