package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pennon Blade — Artifact — Equipment {3} (Rise of the Eldrazi, #5581):
//
//	"Equipped creature gets +1/+1 for each creature you control.
//	 Equip {4} ({4}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Kemba's Banner's count: "you" is the Equipment's controller, and the
// equipped creature counts itself when that player controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b97e6acd-a676-429e-b875-b734759ebea7",
		Name:         "Pennon Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttachedPer(1, 1, effects.RoeB1CreaturesYouControl)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
