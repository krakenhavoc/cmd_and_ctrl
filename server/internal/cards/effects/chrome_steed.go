package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Steed — Artifact Creature — Horse {4}, 2/2 (Scars of Mirrodin,
// #5577):
//
//	"Metalcraft — This creature gets +2/+2 as long as you control three
//	 or more artifacts."
//
// The metalcraft self pump (somB1MetalcraftPump); the Steed counts as one
// of the three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8475f55e-473c-4df0-bae2-3cbbbde4ff12",
		Name:         "Chrome Steed",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB1MetalcraftPump("Metalcraft — This creature gets +2/+2 as long as you control three or more artifacts.", 2, 2),
		},
	})
}
