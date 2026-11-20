package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Adept — Creature — Human Wizard {1}{U}{U}, 2/2 (Magic 2011,
// #5579):
//
//	"When this creature enters, return target creature to its owner's
//	 hand."
//
// Man-o'-War's enters trigger: the target is chosen as the trigger goes
// on the stack (CR 603.3d), any creature, the Adept itself included, and
// a target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e076067c-77b7-4a37-a62d-51434392dd53",
		Name:         "Aether Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Aether Adept — return target creature to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature")),
		},
	})
}
