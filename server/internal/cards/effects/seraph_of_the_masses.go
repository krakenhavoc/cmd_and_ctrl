package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seraph of the Masses — Creature — Angel {5}{W}{W}, */*:
//
//	"Convoke
//	 Flying
//	 Seraph of the Masses's power and toughness are each equal to the
//	 number of creatures you control."
//
// Convoke is Spec.TapCost (Ancient Imperiosaur); the size is Burrowguard
// Mentor's layer 7a characteristic-defining ability (CR 604.3, 613.4a),
// the Seraph itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d6f75f1-dd9c-4d94-aea8-3bdec41eaa29",
		Name:            "Seraph of the Masses",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Convoke(),
		Static: []game.StaticAbility{{
			Label:     "Seraph of the Masses's power and toughness are each equal to the number of creatures you control.",
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
