package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeong Jeong's Deserters — Creature — Human Rebel Ally {1}{W}, 1/2:
//
//	"When this creature enters, put a +1/+1 counter on target
//	 creature."
//
// Any creature, itself included. No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a06d281-3ca2-498f-91d0-e13e7b0869f9",
		Name:         "Jeong Jeong's Deserters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Jeong Jeong's Deserters — put a +1/+1 counter on target creature", rfCreatureFCounterOnTargetCreature),
				TargetCreature("target creature"),
			),
		},
	})
}
