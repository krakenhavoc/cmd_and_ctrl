package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adamaro, First to Desire — Legendary Creature — Spirit {1}{R}{R}, */*
// (Saviors of Kamigawa, #5601):
//
//	"Adamaro's power and toughness are each equal to the number of cards in
//	 the hand of the opponent with the most cards in hand."
//
// A characteristic-defining ability (CR 604.3, layer 7a), read live: the
// biggest hand among Adamaro's controller's opponents still in the game. A
// hand move drops the layer cache while Adamaro is out (DependsOnHandSize).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18f34bfd-7675-49df-91ea-695c510cd3c3",
		Name:         "Adamaro, First to Desire",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:             "Adamaro's power and toughness are each equal to the number of cards in the hand of the opponent with the most cards in hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7A_CDA,
			DependsOnHandSize: true,
			AppliesTo:         effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.OldB592MostCardsInAnOpponentsHand(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
