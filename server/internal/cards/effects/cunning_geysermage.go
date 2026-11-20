package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "746768bc-9abc-495f-ad00-4b58497dce0d",
		Name:          "Cunning Geysermage",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Cunning Geysermage — kicked, return up to one other target creature to its owner's hand",
					bounceChosenTarget),
				Another(UpToOneTargetCreature("up to one other target creature")),
			),
		},
	})
}
