package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeong Jeong's Deserters — Creature — Human Rebel Ally {1}{W}, 1/2:
//
//	"When this creature enters, put a +1/+1 counter on target
//	 creature."
//
// Any creature, itself included. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a06d281-3ca2-498f-91d0-e13e7b0869f9",
		Name:         "Jeong Jeong's Deserters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Jeong Jeong's Deserters — put a +1/+1 counter on target creature", effects.RfCreatureFCounterOnTargetCreature),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
