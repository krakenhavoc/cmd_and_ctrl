package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanguard's Shield — Artifact — Equipment {2}:
//
//	"Equipped creature gets +0/+3 and can block an additional creature
//	 each combat.
//	 Equip {3}"
//
// Bonesplitter's shape (PumpAttached + EquipAbility) plus
// CanBlockAdditional (#1706) over the same AttachedToSource relation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc628018-f5d0-4bd2-91e7-ee4e4a5af9aa",
		Name:         "Vanguard's Shield",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 3),
			effects.CanBlockAdditional(effects.AttachedToSource, 1),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}
