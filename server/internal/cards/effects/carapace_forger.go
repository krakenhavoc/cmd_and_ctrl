package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carapace Forger — Creature — Elf Artificer {1}{G}, 2/2 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature gets +2/+2 as long as you control three
//	 or more artifacts."
//
// Chrome Steed's metalcraft self pump (somB1MetalcraftPump), in layer 7c.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e623e0a-3c2d-4536-b12f-9437ef820499",
		Name:         "Carapace Forger",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			somB1MetalcraftPump("Metalcraft — This creature gets +2/+2 as long as you control three or more artifacts.", 2, 2),
		},
	})
}
