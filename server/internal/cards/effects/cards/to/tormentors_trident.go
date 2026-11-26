package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormentor's Trident — Artifact — Equipment {2}:
//
//	"Equipped creature gets +3/+0 and attacks each combat if able.
//	 Equip {3}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b74992b8-2857-403b-88cf-f2718873571a",
		Name:         "Tormentor's Trident",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.AttacksEachCombatWhere(effects.AttachedToSource),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
