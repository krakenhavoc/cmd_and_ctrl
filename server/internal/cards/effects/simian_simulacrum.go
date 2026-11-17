package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simian Simulacrum — Artifact Creature — Ape {3}, 2/1:
//
//	"When this creature enters, put two +1/+1 counters on target creature
//	 you control.
//	 Unearth {2}{G}{G}"
//
// The target may be the Simulacrum itself. Unearth is the shared keyword
// constructor (unearth.go).
func init() {
	Register(Spec{
		OracleID:     "77eb44ca-53e9-4924-b19f-31732f14ef14",
		Name:         "Simian Simulacrum",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Simian Simulacrum — put two +1/+1 counters on target creature you control",
					rfCreatureFCountersOnTarget(2)),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 0, 0))),
		},
		Activated: []ActivatedAbility{Unearth("{2}{G}{G}")},
	})
}
