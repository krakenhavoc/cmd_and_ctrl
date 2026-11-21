package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Short Bow — Artifact — Equipment {2} (Bloomburrow, #4467):
//
//	"Equipped creature gets +1/+1 and has reach and vigilance.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Behemoth Sledge's two statics with smaller numbers and other keywords,
// and the shared equip ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce5f5ff9-13f7-41eb-ada0-7f73de42acef",
		Name:         "Short Bow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("reach", "vigilance"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
