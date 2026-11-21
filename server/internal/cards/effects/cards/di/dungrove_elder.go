package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dungrove Elder — Creature — Treefolk {2}{G}, */*:
//
//	"Hexproof
//	 Dungrove Elder's power and toughness are each equal to the number of
//	 Forests you control."
//
// A layer 7a characteristic-defining ability counting lands with the
// Forest land type, not lands named Forest, read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32083c9f-04c6-4b6f-9e24-cc8f3cae3e3e",
		Name:            "Dungrove Elder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{{
			Label:     "Dungrove Elder's power and toughness are each equal to the number of Forests you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Forest"))
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
