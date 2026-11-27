package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meishin, the Mind Cage — Legendary Enchantment {4}{U}{U}{U} (Saviors of
// Kamigawa, #5601):
//
//	"All creatures get -X/-0, where X is the number of cards in your hand."
//
// A layer 7c modification (CR 613.4c) of every creature, whoever controls it,
// read live from Meishin's controller's hand, so it follows every draw and
// discard (DependsOnHandSize drops the layer cache when a hand moves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b8eea0b-6f8c-449c-9ab8-a22a035465d6",
		Name:         "Meishin, the Mind Cage",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:             "All creatures get -X/-0, where X is the number of cards in your hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power -= effects.CardsInHand(g, source.Controller)
			},
		}},
	})
}
