package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spidersilk Net — Artifact — Equipment {0} (Zendikar, #5583):
//
//	"Equipped creature gets +0/+2 and has reach. (It can block
//	 creatures with flying.)
//	 Equip {2} ({2}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The shared attachment pump and keyword grant, and the sorcery-speed
// EquipAbility.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04979f18-1ba8-47a3-a078-ab6b84bfac72",
		Name:         "Spidersilk Net",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.GrantToAttached("reach"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
