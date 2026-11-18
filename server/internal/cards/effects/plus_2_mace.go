package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// +2 Mace — Artifact — Equipment {1}{W} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +2/+2.
//	Equip {3}
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "629fe1be-272d-465f-b9b1-2ce177410f13",
		Name:         "+2 Mace",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Activated:    []ActivatedAbility{EquipAbility("{3}")},
	})
}
