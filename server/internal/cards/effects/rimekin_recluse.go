package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rimekin Recluse — Creature — Elemental Wizard {2}{U}, 3/2 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, return up to one other target creature to
//	 its owner's hand."
//
// Matterbending Mage's enter trigger: up to one creature other than the
// Recluse, any controller's, chosen as the trigger goes on the stack;
// with no target it does nothing and a target that left is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad99e430-e13f-4819-a2ee-bf3332ef517d",
		Name:         "Rimekin Recluse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rimekin Recluse — return up to one other target creature to its owner's hand", bounceChosenTarget),
				Another(UpToOneTargetCreature("up to one other target creature")),
			),
		},
	})
}
