package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auriok Sunchaser — Creature — Human Soldier {1}{W}, 1/1 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — As long as you control three or more artifacts, this
//	 creature gets +2/+2 and has flying."
//
// Two statics on the one metalcraft gate: the pump in layer 7c
// (somB1MetalcraftPump) and the flying in layer 6
// (somB2MetalcraftKeywords). Both turn on and off together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e07d314-119f-4918-9105-77543ccae7e6",
		Name:         "Auriok Sunchaser",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB1MetalcraftPump("Metalcraft — As long as you control three or more artifacts, this creature gets +2/+2.", 2, 2),
			somB2MetalcraftKeywords("Metalcraft — As long as you control three or more artifacts, this creature has flying.", "flying"),
		},
	})
}
