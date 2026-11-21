package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Towering Gibbon — Creature — Ape {3}{G}, */4 (Jumpstart 2022, #5948):
//
//	"Reach
//	 Towering Gibbon's power is equal to the greatest mana value among
//	 creatures you control."
//
// A characteristic-defining ability (layer 7a), read live, the Gibbon itself
// counted. A token has mana value 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0684771f-46ea-40a3-b021-e98a7d7cd79d",
		Name:            "Towering Gibbon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Label:     "Towering Gibbon's power is equal to the greatest mana value among creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.GapB153GreatestManaValueAmongYourCreatures(g, source.Controller)
			},
		}},
	})
}
