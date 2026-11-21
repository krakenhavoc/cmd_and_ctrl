package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snapsail Glider — Artifact Creature — Construct {3}, 2/2 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature has flying as long as you control three
//	 or more artifacts."
//
// The metalcraft self grant (somB2MetalcraftKeywords), in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43e2e1ed-5c6a-49b4-8ab8-57861c02fc3e",
		Name:         "Snapsail Glider",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB2MetalcraftKeywords("Metalcraft — This creature has flying as long as you control three or more artifacts.", "flying"),
		},
	})
}
