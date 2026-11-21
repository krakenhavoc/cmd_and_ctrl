package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Missionary — Creature — Phyrexian Human Cleric {1}{W}, 2/3
// (Dominaria United, #5487):
//
//	"Kicker {1}{B} (You may pay an additional {1}{B} as you cast this
//	 spell.)
//	 Lifelink
//	 When this creature enters, if it was kicked, return target creature
//	 card from your graveyard to your hand."
//
// Along the Crooked Way's graveyard-to-hand return behind the kicked
// intervening "if" (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be04f54f-59fa-4f0a-b8b5-790d568932b6",
		Name:            "Phyrexian Missionary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Phyrexian Missionary — kicked: return target creature card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn())),
		},
	})
}
