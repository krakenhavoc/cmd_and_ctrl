package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surrakar Banisher — Creature — Surrakar {4}{U}, 3/3 (Worldwake, #5582):
//
//	"When this creature enters, you may return target tapped creature to
//	 its owner's hand."
//
// Harbinger of the Tides' enters trigger, with any tapped creature as
// the target: chosen as the trigger goes on the stack (CR 603.3d), the
// "you may" asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d48fd4b7-b60c-4932-abed-a8df1184db25",
		Name:         "Surrakar Banisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Surrakar Banisher — return target tapped creature to its owner's hand", bounceTheTargetToHand),
				TargetCreature("target tapped creature", hobB3TappedCreature())),
				"Surrakar Banisher — return target tapped creature to its owner's hand?"),
		},
	})
}
