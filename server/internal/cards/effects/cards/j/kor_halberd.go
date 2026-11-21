package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Halberd — Artifact — Equipment {W} (March of the Machine):
//
//	"Equipped creature gets +1/+1 and has vigilance.
//	 Equip {1}"
//
// PumpAttached and GrantToAttached, as on Batterbone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "450e84c7-afbe-4b30-a3e9-1dff968a5200",
		Name:         "Kor Halberd",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
