package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murasa Behemoth — Creature — Beast {4}{G}{G}, 5/5 (Modern Horizons, #5519):
//
//	"Trample
//	 This creature gets +3/+3 as long as there is a land card in your
//	 graveyard."
//
// A layer-7c self static read on every recompute off your graveyard (a
// zone move invalidates the layers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0dfdeaa0-fa12-49cf-b361-efa49d3ffcb2",
		Name:            "Murasa Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "Gets +3/+3 as long as there is a land card in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.B17LandCardsInGraveyard(g, source.Controller) > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
				c.Toughness += 3
			},
		}},
	})
}
