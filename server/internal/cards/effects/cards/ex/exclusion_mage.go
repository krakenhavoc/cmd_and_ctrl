package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fd28fe55-36cd-4242-8dd4-edb58fbb9895",
		Name:         "Exclusion Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Exclusion Mage — return target creature an opponent controls to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
