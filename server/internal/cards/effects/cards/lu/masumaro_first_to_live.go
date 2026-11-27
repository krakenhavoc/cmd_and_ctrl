package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Masumaro, First to Live — Legendary Creature — Spirit {3}{G}{G}{G}, */*
// (Saviors of Kamigawa, #5601):
//
//	"Masumaro's power and toughness are each equal to twice the number of
//	 cards in your hand."
//
// A characteristic-defining ability (CR 604.3, layer 7a), read live from
// its controller's hand (DependsOnHandSize drops the layer cache when a hand
// moves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8af347b8-89fa-499e-8e5a-77463edd6773",
		Name:         "Masumaro, First to Live",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:             "Masumaro's power and toughness are each equal to twice the number of cards in your hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7A_CDA,
			DependsOnHandSize: true,
			AppliesTo:         effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2 * effects.CardsInHand(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
