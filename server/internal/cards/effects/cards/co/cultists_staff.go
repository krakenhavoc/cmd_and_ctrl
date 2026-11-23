package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cultist's Staff — Artifact — Equipment {2} (Eldritch Moon, #5541):
//
//	Equipped creature gets +2/+2.
//	Equip {3}
//
// +2 Mace's two lines.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e211792-2656-4bee-80ac-60eac0c45c2d",
		Name:         "Cultist's Staff",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
