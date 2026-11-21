package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d48fd4b7-b60c-4932-abed-a8df1184db25",
		Name:         "Surrakar Banisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Surrakar Banisher — return target tapped creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target tapped creature", effects.HobB3TappedCreature())),
				"Surrakar Banisher — return target tapped creature to its owner's hand?"),
		},
	})
}
