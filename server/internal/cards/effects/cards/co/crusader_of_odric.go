package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crusader of Odric — Creature — Human Soldier {2}{W}, */* (Foundations,
// #4465):
//
//	"Crusader of Odric's power and toughness are each equal to the number
//	 of creatures you control."
//
// Burrowguard Mentor's characteristic-defining ability (layer 7a), read
// live, the Crusader itself counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea384b0d-3091-4d50-b15f-1f6763647b7c",
		Name:         "Crusader of Odric",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Crusader of Odric's power and toughness are each equal to the number of creatures you control.",
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
