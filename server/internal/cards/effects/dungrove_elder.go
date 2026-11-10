package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "32083c9f-04c6-4b6f-9e24-cc8f3cae3e3e",
		Name:            "Dungrove Elder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{{
			Label:     "Dungrove Elder's power and toughness are each equal to the number of Forests you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Forest"))
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
