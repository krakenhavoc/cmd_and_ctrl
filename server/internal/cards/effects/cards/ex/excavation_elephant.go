package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "45fa04ee-3b5c-45b3-8c09-bda08cca3630",
		Name:          "Excavation Elephant",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Excavation Elephant — kicked: return target artifact card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
		},
	})
}
