package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Scimitar — Artifact — Equipment {1} (Foundations Jumpstart):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cde26d69-f3e7-4dd0-a53b-cd0ec812d717",
		Name:         "Leonin Scimitar",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
