package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3cc333a1-c854-4b7c-8002-910e48222371",
		Name:         "Well-Worn Spatula",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Well-Worn Spatula — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
