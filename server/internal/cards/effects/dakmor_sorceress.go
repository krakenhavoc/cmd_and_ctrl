package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dakmor Sorceress — Creature — Human Wizard Sorcerer {5}{B}, */4
// (Mystery Booster 2):
//
//	"Dakmor Sorceress's power is equal to the number of Swamps you control."
//
// A characteristic-defining ability (layer 7a, CR 604.3), counting the
// controller's Swamps through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be5d9522-40b4-4f45-978b-d78bd00f9491",
		Name:         "Dakmor Sorceress",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of Swamps you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = swampsControlledBy(g, source.Controller, "Swamp")
			},
		}},
	})
}
