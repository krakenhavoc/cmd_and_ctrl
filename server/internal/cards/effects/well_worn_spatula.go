package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Well-Worn Spatula — Artifact — Equipment {1} (The Hobbit, HOB):
//
//	"When this Equipment enters, you gain 2 life.
//	 Equipped creature gets +1/+1.
//	 Equip {1}"
//
// The ETB is a trigger on the Equipment's own entry. The pump is
// PumpAttached and equip is the shared EquipAbility (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cc333a1-c854-4b7c-8002-910e48222371",
		Name:         "Well-Worn Spatula",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Well-Worn Spatula — you gain 2 life", Do(GainLife{Amount: 2})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
