package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanquisher's Axe — Artifact — Equipment {1} (Dominaria United, S82):
//
//	"Equipped creature gets +2/+0.
//	 Equip {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbdd3ba9-d799-476b-87bc-a689d8e3cff9",
		Name:         "Vanquisher's Axe",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
