package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auriok Edgewright — Creature — Human Soldier {W}{W}, 2/2 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature has double strike as long as you control three
//	 or more artifacts."
//
// The metalcraft self grant (somB2MetalcraftKeywords), in layer 6.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b718639d-cfba-4430-b5e3-1e644edd36d7",
		Name:         "Auriok Edgewright",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB2MetalcraftKeywords("Metalcraft — This creature has double strike as long as you control three or more artifacts.", "double strike"),
		},
	})
}
