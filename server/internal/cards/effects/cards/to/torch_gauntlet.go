package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch Gauntlet — Artifact — Equipment {2} (Kaladesh, #5539):
//
//	"Equipped creature gets +2/+0.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Bat'leth's static and equip: a +2/+0 on the equipped creature and
// equip {2} at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07fdc0b0-7131-411d-b0a8-f667165a64b8",
		Name:         "Torch Gauntlet",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
