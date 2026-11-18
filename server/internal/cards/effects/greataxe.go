package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greataxe — Artifact — Equipment {1} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +4/+0.
//	Equip {5}
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "843f3373-b8ef-416c-9163-aa1aa6b4a51d",
		Name:         "Greataxe",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(4, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{5}")},
	})
}
