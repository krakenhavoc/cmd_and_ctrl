package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supply Runners — Creature — Dog {4}{W}, 2/2 (Commander Masters,
// #5948):
//
//	"When this creature enters, put a +1/+1 counter on each other
//	 creature you control."
//
// The set of creatures is fixed as the trigger resolves; Supply Runners
// itself gets none (it says "other").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d6da25d-7b62-4a64-bbaa-ae8edab1dd51",
		Name:         "Supply Runners",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Supply Runners — put a +1/+1 counter on each other creature you control",
				putCounterOnEachOtherCreatureYouControl),
		},
	})
}
