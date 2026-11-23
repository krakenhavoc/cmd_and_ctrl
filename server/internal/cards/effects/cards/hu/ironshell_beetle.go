package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironshell Beetle — Creature — Insect {1}{G}, 1/1 (Jumpstart):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Duskshell Crawler's entry counter without its trample grant: any
// creature, the Beetle itself included, picked as the trigger goes on
// the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "693ca077-2b91-4de1-8136-8fe45f8ea5a7",
		Name:         "Ironshell Beetle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ironshell Beetle — put a +1/+1 counter on target creature", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
