package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Esgaroth Garrison — Creature — Human Soldier {4}{W}, */5:
//
//	"Esgaroth Garrison's power is equal to the number of creatures you
//	 control.
//	 When this creature enters, recruit. (Draw a card, then discard a
//	 card. If you discarded a nonland card, create a 1/1 white Human
//	 Soldier creature token.)"
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a),
// Adeline's shape: it counts the Garrison itself, and a Soldier the
// recruit makes the moment it arrives.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:     "84597236-2cfc-4654-b126-a8ee252f4d14",
		Name:         "Esgaroth Garrison",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = b04CreaturesControlled(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Esgaroth Garrison — recruit", Do(Recruit{})),
		},
	})
}
