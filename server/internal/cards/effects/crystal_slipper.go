package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Slipper — Artifact — Equipment {1}{R} (Throne of Eldraine, #5516):
//
//	"Equipped creature gets +1/+0 and has haste.
//	 Equip {1}"
//
// Boots of Speed's text at a red cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c2d2888-38be-41f9-aaf2-8337e87d2aed",
		Name:         "Crystal Slipper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0), GrantToAttached("haste")},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
