package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slag Fiend — Creature — Phyrexian Construct {R}, */*:
//
//	"Slag Fiend's power and toughness are each equal to the number of artifact
//	 cards in all graveyards."
//
// A characteristic-defining ability (CR 604.3, layer 7a), recomputed as the
// graveyards change.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c5e38ca-39d7-4e27-8bb7-687d86fd67c4",
		Name:         "Slag Fiend",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Slag Fiend's power and toughness are each equal to the number of artifact cards in all graveyards.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.ArtifactCardsInAllGraveyards(g)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
