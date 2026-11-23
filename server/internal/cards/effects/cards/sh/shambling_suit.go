package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shambling Suit — Artifact Creature — Construct {3}, */3 (Jumpstart 2022):
//
//	"Shambling Suit's power is equal to the number of artifacts and/or
//	 enchantments you control."
//
// A layer 7a characteristic-defining ability (CR 604.3, 613.4a), counted as
// the layers run over every artifact or enchantment its controller has on
// the battlefield, the Suit itself included. An artifact enchantment counts
// once. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "862b777a-4816-4a3e-be8d-ddfcb72268e8",
		Name:         "Shambling Suit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Shambling Suit's power is equal to the number of artifacts and/or enchantments you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.Controller == source.Controller && (p.IsArtifact() || p.IsEnchantment()) {
						n++
					}
				}
				c.Power = n
			},
		}},
	})
}
