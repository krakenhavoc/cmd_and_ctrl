package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warrior's Sword — Artifact — Equipment {3}{R} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +3/+2 and is a Warrior in addition to its
//	 other types.
//	 Equip {5}"
//
// The job-select Equipment shape (Bard's Bow): the shared JobSelect
// trigger, a layer 7c pump and a layer 4 creature-type addition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a339961e-3555-467f-a4f8-c5eaac7e7ed1",
		Name:         "Warrior's Sword",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 2),
			AttachedIsAlso("Warrior"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Warrior's Sword")},
		Activated: []ActivatedAbility{EquipAbility("{5}")},
	})
}
