package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chitinous Cloak — Artifact — Equipment {3} (Oath of the Gatewatch, #5545):
//
//	"Equipped creature gets +2/+2 and has menace. (It can't be blocked
//	 except by two or more creatures.)
//	 Equip {3}"
//
// Arm-Mounted Anchor's two statics on the equipped creature, and
// Behemoth Sledge's equip row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5487098-fd24-49af-9634-4827a5c2eac4",
		Name:         "Chitinous Cloak",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("menace"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
