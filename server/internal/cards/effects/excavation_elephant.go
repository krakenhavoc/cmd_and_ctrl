package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excavation Elephant — Creature — Elephant {4}{W}, 3/5 (Dominaria, #5530):
//
//	"Kicker {1}{W} (You may pay an additional {1}{W} as you cast this spell.)
//	 When this creature enters, if it was kicked, return target artifact card from your graveyard to your hand."
//
// Phyrexian Missionary's kicked return, for an artifact card. The
// kicker is read off the permanent (CR 603.4); a reanimated Elephant was
// not kicked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "45fa04ee-3b5c-45b3-8c09-bda08cca3630",
		Name:          "Excavation Elephant",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Excavation Elephant — kicked: return target artifact card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return returnLegalGraveyardTargetsToHand(NewContext(g, item))
				}),
				TargetCardInGraveyard("target artifact card from your graveyard", Artifact(), YouOwn())),
		},
	})
}
