package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0684771f-46ea-40a3-b021-e98a7d7cd79d",
		Name:            "Towering Gibbon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Label:     "Towering Gibbon's power is equal to the greatest mana value among creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = gapB153GreatestManaValueAmongYourCreatures(g, source.Controller)
			},
		}},
	})
}
