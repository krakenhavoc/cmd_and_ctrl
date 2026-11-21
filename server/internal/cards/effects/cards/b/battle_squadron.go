package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Squadron — Creature — Goblin {3}{R}{R}, */* (Jumpstart 2022,
// #5948):
//
//	"Flying
//	 Battle Squadron's power and toughness are each equal to the number of
//	 creatures you control."
//
// Crusader of Odric's characteristic-defining ability (layer 7a), read live,
// the Squadron itself counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5783e71e-2497-4241-9d1a-8bced4745b1d",
		Name:            "Battle Squadron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Battle Squadron's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
