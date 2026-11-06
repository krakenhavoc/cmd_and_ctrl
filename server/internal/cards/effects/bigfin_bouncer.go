package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bigfin Bouncer — Creature — Shark Pirate {3}{U}, 3/2 (Foundations,
// #4465):
//
//	"When this creature enters, return target creature an opponent
//	 controls to its owner's hand."
//
// Pelia, Immortal Innovator's enter trigger: the target is re-checked
// as it resolves (CR 608.2b), and an illegal one does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e473298-0b67-42b8-90a8-3bcdd73da473",
		Name:         "Bigfin Bouncer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Bigfin Bouncer — return target creature an opponent controls to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
