package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crusader of Odric — Creature — Human Soldier {2}{W}, */* (Foundations,
// #4465):
//
//	"Crusader of Odric's power and toughness are each equal to the number
//	 of creatures you control."
//
// Burrowguard Mentor's characteristic-defining ability (layer 7a), read
// live, the Crusader itself counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea384b0d-3091-4d50-b15f-1f6763647b7c",
		Name:         "Crusader of Odric",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Crusader of Odric's power and toughness are each equal to the number of creatures you control.",
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
