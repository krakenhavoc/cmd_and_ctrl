package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pit Keeper — {1}{B} Creature — Human Wizard 2/1 (Jumpstart 2022, #5948):
//
//	"When this creature enters, if you have four or more creature cards in
//	 your graveyard, you may return target creature card from your
//	 graveyard to your hand."
//
// The intervening "if" (CR 603.4) is checked when it enters, so with
// fewer than four there is no trigger and no target, and again as it
// resolves. Gravedigger's optional return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1c5bfae-07cb-4194-8a86-201503e17085",
		Name:         "Pit Keeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, pitKeeperFourCreatureCards),
					"Pit Keeper — return target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						if effects.B11CreatureCardsInGraveyard(g, item.Controller) < 4 {
							return nil
						}
						return effects.ReturnTargetCardToHand(g, item)
					}),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
				"Pit Keeper — return it to your hand?"),
		},
	})
}

func pitKeeperFourCreatureCards(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.B11CreatureCardsInGraveyard(g, source.Controller) >= 4
}
