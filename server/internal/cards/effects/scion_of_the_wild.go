package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scion of the Wild — Creature — Avatar {1}{G}{G}, */* (Ravnica: City of
// Guilds, #5600):
//
//	"Scion of the Wild's power and toughness are each equal to the number
//	 of creatures you control."
//
// Crusader of Odric's characteristic-defining ability (CR 604.3, layer
// 7a): it counts itself, and it works in every zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8aaa2217-012e-4e34-ba44-4986d7ee9ece",
		Name:         "Scion of the Wild",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Scion of the Wild's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
