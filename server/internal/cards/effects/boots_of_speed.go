package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boots of Speed — Artifact — Equipment {R} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +1/+0 and has haste.
//	Equip {1}
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f6601bc-cb12-4054-9027-e4a0ebdae1a3",
		Name:         "Boots of Speed",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0), GrantToAttached("haste")},
		Activated:    []ActivatedAbility{EquipAbility("{1}")},
	})
}
