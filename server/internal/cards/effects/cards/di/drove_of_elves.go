package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drove of Elves — Creature — Elf {3}{G} */* (Commander 2014):
//
//	"Hexproof
//	 Drove of Elves's power and toughness are each equal to the number of
//	 green permanents you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a), counting every
// green permanent its controller has, itself included, read through the
// layered colours.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ba81fbe-2c53-495b-9a2b-f124475bd71b",
		Name:            "Drove of Elves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{{
			Label:     "Drove of Elves's power and toughness are each equal to the number of green permanents you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CountPermanentsYouControl(g, source, func(p *game.Card) bool { return p.HasColor(source.TextColor("G")) })
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
