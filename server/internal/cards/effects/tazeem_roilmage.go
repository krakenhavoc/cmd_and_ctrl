package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tazeem Roilmage — Creature — Merfolk Wizard {1}{U}, 2/1 (Zendikar
// Rising, #5508):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 When this creature enters, if it was kicked, return target instant
//	 or sorcery card from your graveyard to your hand."
//
// The kicked enters trigger is Heartstabber Mosquito's intervening if
// (CR 603.4); the return is Ardent Elementalist's, a target in your own
// graveyard chosen as the trigger goes on the stack and checked again as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "55a277b4-575f-463b-913c-7857acad20be",
		Name:          "Tazeem Roilmage",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{4}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Tazeem Roilmage — kicked, return target instant or sorcery card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
