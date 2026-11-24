package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodrage Brawler — Creature — Minotaur Warrior {1}{R}, 4/3 (Jumpstart,
// #5510):
//
//	"When this creature enters, discard a card."
//
// Mandatory, and the controller chooses the card (CR 701.9a); with an empty
// hand nothing is discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8603e2d4-93f3-491a-91e7-a187fc6db599",
		Name:         "Bloodrage Brawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bloodrage Brawler — discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.DiscardOne(g, item, item.Controller)
			}),
		},
	})
}
