package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simian Simulacrum — Artifact Creature — Ape {3}, 2/1:
//
//	"When this creature enters, put two +1/+1 counters on target creature
//	 you control.
//	 Unearth {2}{G}{G}"
//
// The target may be the Simulacrum itself. Unearth is the shared keyword
// constructor (unearth.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77eb44ca-53e9-4924-b19f-31732f14ef14",
		Name:         "Simian Simulacrum",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Simian Simulacrum — put two +1/+1 counters on target creature you control",
					effects.RfCreatureFCountersOnTarget(2)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{G}{G}")},
	})
}
