package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bc628018-f5d0-4bd2-91e7-ee4e4a5af9aa",
		Name:         "Vanguard's Shield",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(0, 3),
			CanBlockAdditional(AttachedToSource, 1),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
