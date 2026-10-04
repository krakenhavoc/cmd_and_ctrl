package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Ten Rings — Legendary Artifact {8}:
//
//	"Your maximum hand size is ten.
//	 At the beginning of your end step, if you have fewer than ten
//	 cards in hand, draw cards equal to the difference."
//
// The maximum is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order. The end-step draw is
// AtYourEndStepDrawUpTo, shared with Doctor Octopus, Master Planner.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a82d855c-50c8-43f2-99ba-a84fe84539c5",
		Name:         "The Ten Rings",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{YourMaxHandSizeIs(10)},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepDrawUpTo("The Ten Rings", 10),
		},
	})
}
