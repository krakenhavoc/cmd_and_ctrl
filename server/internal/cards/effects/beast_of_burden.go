package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast of Burden — Artifact Creature — Golem {6}, */*:
//
//	"Beast of Burden's power and toughness are each equal to the
//	 number of creatures on the battlefield."
//
// A layer 7a characteristic-defining ability over every creature on
// the battlefield, whoever controls it, itself included. No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "67389ccc-dacc-4ce3-a9a2-f551a0d6e9d3",
		Name:         "Beast of Burden",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.IsCreature() {
						n++
					}
				}
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
