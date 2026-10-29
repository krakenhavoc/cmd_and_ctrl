package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirkwood Pathmaker — Creature — Elf Ranger {2}{G}, */* (The Hobbit,
// HOB):
//
//	"Mirkwood Pathmaker's power and toughness are each equal to the
//	 number of lands you control."
//
// Lumra's size without the keywords or the trigger: a
// characteristic-defining ability (CR 604.3) in layer 7a that sets both
// values to the land count on every recompute (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1b36dde-9cde-46c8-8d5a-9ad93feba41f",
		Name:         "Mirkwood Pathmaker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{hobA2LandsPT()},
	})
}

// hobA2LandsPT is "this creature's power and toughness are each equal to
// the number of lands you control" as a layer 7a CDA.
func hobA2LandsPT() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Power and toughness are each equal to the number of lands you control.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7A_CDA,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := b10LandsControlled(g, source.Controller)
			c.Power = n
			c.Toughness = n
		},
	}
}
