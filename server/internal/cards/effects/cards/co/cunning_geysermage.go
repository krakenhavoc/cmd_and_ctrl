package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Geysermage — Creature — Human Wizard {2}{U}, 3/2 (Zendikar
// Rising, #5508):
//
//	"Kicker {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell.)
//	 When this creature enters, if it was kicked, return up to one other
//	 target creature to its owner's hand."
//
// The kicked enters trigger is Heartstabber Mosquito's intervening if
// (CR 603.4); the bounce is Matterbending Mage's, up to one creature
// other than the Geysermage, any controller's. With no target it does
// nothing, and a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "746768bc-9abc-495f-ad00-4b58497dce0d",
		Name:          "Cunning Geysermage",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Cunning Geysermage — kicked, return up to one other target creature to its owner's hand",
					effects.BounceChosenTarget),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature")),
			),
		},
	})
}
