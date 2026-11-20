package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honed Khopesh — Artifact — Equipment {1}
// (Amonkhet, #5535):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"
//
// The shared pump and equip constructors.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2142c41-ff66-4276-b426-fd945ed9b0f6",
		Name:         "Honed Khopesh",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
