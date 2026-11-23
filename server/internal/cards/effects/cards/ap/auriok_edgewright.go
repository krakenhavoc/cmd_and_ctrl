package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Edgewright — Creature — Human Soldier {W}{W}, 2/2 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This creature has double strike as long as you control three
//	 or more artifacts."
//
// The metalcraft self grant (somB2MetalcraftKeywords), in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b718639d-cfba-4430-b5e3-1e644edd36d7",
		Name:         "Auriok Edgewright",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB2MetalcraftKeywords("Metalcraft — This creature has double strike as long as you control three or more artifacts.", "double strike"),
		},
	})
}
