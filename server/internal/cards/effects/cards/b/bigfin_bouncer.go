package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4e473298-0b67-42b8-90a8-3bcdd73da473",
		Name:         "Bigfin Bouncer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Bigfin Bouncer — return target creature an opponent controls to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
