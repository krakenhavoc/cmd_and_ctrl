package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aeronaut's Wings — Artifact — Equipment {2} (The Brothers' War):
//
//	"Equipped creature gets +1/+0 and has flying.
//	 Equip {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3bad2e2-1024-4261-b3d7-82a6998416fd",
		Name:         "Aeronaut's Wings",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("flying"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
