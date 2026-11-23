package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ezuri's Brigade — Creature — Elf Warrior {2}{G}{G}, 4/4 (Scars of Mirrodin, #5577):
//
//	"Metalcraft — As long as you control three or more artifacts, this
//	 creature gets +4/+4 and has trample."
//
// Two statics on the one metalcraft gate: the pump in layer 7c
// (somB1MetalcraftPump) and the trample in layer 6
// (somB2MetalcraftKeywords). Both turn on and off together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae32c0a5-0874-4cab-9a24-29fa572f3193",
		Name:         "Ezuri's Brigade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB1MetalcraftPump("Metalcraft — As long as you control three or more artifacts, this creature gets +4/+4.", 4, 4),
			effects.SomB2MetalcraftKeywords("Metalcraft — As long as you control three or more artifacts, this creature has trample.", "trample"),
		},
	})
}
