package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snapsail Glider — Artifact Creature — Construct {3}, 2/2 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature has flying as long as you control three
//	 or more artifacts."
//
// The metalcraft self grant (somB2MetalcraftKeywords), in layer 6.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43e2e1ed-5c6a-49b4-8ab8-57861c02fc3e",
		Name:         "Snapsail Glider",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB2MetalcraftKeywords("Metalcraft — This creature has flying as long as you control three or more artifacts.", "flying"),
		},
	})
}
