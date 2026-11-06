package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exclusion Mage — Creature — Human Wizard {2}{U}, 2/2 (Foundations,
// #4465):
//
//	"When this creature enters, return target creature an opponent
//	 controls to its owner's hand."
//
// Bigfin Bouncer's enters trigger on a smaller body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd28fe55-36cd-4242-8dd4-edb58fbb9895",
		Name:         "Exclusion Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Exclusion Mage — return target creature an opponent controls to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
