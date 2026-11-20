package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corridor Monitor — Artifact Creature — Construct {1}{U}, 1/4 (Throne of Eldraine, #5516):
//
//	"When this creature enters, untap target artifact or creature you control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-read at resolution; the untap is Aphetto Alchemist's untapTheTarget.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07b6ee55-a77e-4ba7-9ab7-8a8375c7acc9",
		Name:         "Corridor Monitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Corridor Monitor — untap target artifact or creature you control", untapTheTarget),
				TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl())),
		},
	})
}
