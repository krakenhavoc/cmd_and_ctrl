package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghalma's Warden — Creature — Elephant Soldier {3}{W}, 2/4 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature gets +2/+2 as long as you control three
//	 or more artifacts."
//
// Chrome Steed's metalcraft self pump (somB1MetalcraftPump), in layer 7c.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b17dc39-a564-4c61-b483-98ed3da4361b",
		Name:         "Ghalma's Warden",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB1MetalcraftPump("Metalcraft — This creature gets +2/+2 as long as you control three or more artifacts.", 2, 2),
		},
	})
}
