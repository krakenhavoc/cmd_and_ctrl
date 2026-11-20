package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran's Sidearm — Artifact — Equipment {2} (Magic Origins, #5549):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"
//
// Leonin Scimitar's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f038c11e-b070-40fd-9d0f-e031f7bc2bd2",
		Name:         "Veteran's Sidearm",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
