package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardent Recruit — Creature — Human Soldier {W}, 1/1 (Mirrodin Besieged,
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
		OracleID:     "9275108c-7eaf-4edd-aa6d-e5263d5ac7fd",
		Name:         "Ardent Recruit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB1MetalcraftPump("Metalcraft — This creature gets +2/+2 as long as you control three or more artifacts.", 2, 2),
		},
	})
}
