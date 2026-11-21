package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razorfield Rhino — Artifact Creature — Rhino {6}, 4/4 (Mirrodin Besieged,
// #5576):
//
//	"Metalcraft — This creature gets +2/+2 as long as you control three
//	 or more artifacts."
//
// Chrome Steed's metalcraft self pump (somB1MetalcraftPump): a layer 7c
// +2/+2 that is on only while its controller controls three or more
// artifacts (CR 611.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "834efbf7-7b93-42ac-ab20-0b7b8ca58a16",
		Name:         "Razorfield Rhino",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB1MetalcraftPump("Metalcraft — This creature gets +2/+2 as long as you control three or more artifacts.", 2, 2),
		},
	})
}
