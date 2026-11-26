package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awakened Amalgam — Artifact Creature — Golem {4}, */*:
//
//	"Awakened Amalgam's power and toughness are each equal to the number of
//	 differently named lands you control."
//
// A characteristic-defining ability (CR 604.3) in layer 7a, counting the
// distinct names among the lands its controller controls as the layers are
// applied, so it follows lands coming and going; two lands with one name count
// once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "909d7e77-ad7c-4af2-88db-f22442cfa531",
		Name:         "Awakened Amalgam",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Awakened Amalgam's power and toughness are each equal to the number of differently named lands you control.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB220DifferentlyNamedLands(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
