package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosthetic Injector — Artifact — Equipment {1}
// (Phyrexia: All Will Be One, #5479):
//
//	"Equipped creature gets +0/+2 and has toxic 1.
//	 Equip {1}"
//
// The toxic is cumulative (CR 702.164b) with any the creature has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8026533d-c0e7-48b7-9e1a-9a370899aa57",
		Name:         "Prosthetic Injector",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(0, 2),
			GrantToAttached("toxic 1"),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
