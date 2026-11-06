package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monk's Fist — Artifact — Equipment {2} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+0 and is a Monk in addition to its other
//	 types.
//	 Equip {2}"
//
// Bard's Bow's job-select shape with no keyword: the shared JobSelect
// trigger, a layer 7c pump and a layer 4 creature-type addition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b524149-340e-4f94-9520-6a5258fabb00",
		Name:         "Monk's Fist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			AttachedIsAlso("Monk"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Monk's Fist")},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
