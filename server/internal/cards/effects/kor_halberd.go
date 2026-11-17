package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Halberd — Artifact — Equipment {W} (March of the Machine):
//
//	"Equipped creature gets +1/+1 and has vigilance.
//	 Equip {1}"
//
// PumpAttached and GrantToAttached, as on Batterbone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "450e84c7-afbe-4b30-a3e9-1dff968a5200",
		Name:         "Kor Halberd",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("vigilance"),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
