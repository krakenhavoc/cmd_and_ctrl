package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Henchbots — {4} Artifact Creature — Robot 2/3 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this creature enters, exile target tapped creature an opponent
//	 controls until this creature leaves the battlefield."
//
// Brutal Cathar's exile-until-this-leaves (CR 610.3) with a tapped
// creature as the target. The target must still be tapped as the trigger
// resolves (CR 608.2b); a Henchbots gone before then exiles nothing, and
// the exiled card returns under its owner's control as Henchbots leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f097c91-6c2e-410f-8c8f-c021f39491af",
		Name:         "Henchbots",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetCreature("target tapped creature an opponent controls", tappedPermanent(), OpponentControls()),
			Key:       "Henchbots — exile target tapped creature an opponent controls until this creature leaves the battlefield",
			Effect: exileChosenTargetUntilThisLeaves(
				"Henchbots — the exiled creature returns when Henchbots leaves the battlefield"),
		}},
	})
}
