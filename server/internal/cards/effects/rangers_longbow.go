package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranger's Longbow — Artifact — Equipment {1}{G} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +2/+1 and has reach.
//	Equip {3}
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b9abe7f-050f-4d0a-bc01-f442164f95de",
		Name:         "Ranger's Longbow",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1), GrantToAttached("reach")},
		Activated:    []ActivatedAbility{EquipAbility("{3}")},
	})
}
