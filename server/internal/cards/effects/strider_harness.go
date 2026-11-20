package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strider Harness — Artifact — Equipment {3} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +1/+1 and has haste.
//	 Equip {1}"
//
// Boots of Speed's shape: PumpAttached, GrantToAttached and the equip
// row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd4ad383-8c40-4dda-bc9c-ba8a001d6882",
		Name:         "Strider Harness",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1), GrantToAttached("haste")},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
