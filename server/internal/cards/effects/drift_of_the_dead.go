package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drift of the Dead — Creature — Wall {3}{B}, */* (Ice Age, #5636):
//
//	"Defender
//	 Drift of the Dead's power and toughness are each equal to the
//	 number of snow lands you control."
//
// Defender is a printed keyword (CR 702.3b). The size is Dungrove
// Elder's layer 7a characteristic-defining ability (CR 604.3, 613.4a)
// over lands with the snow supertype (CR 205.4g) its controller
// controls, counted live, so with no snow land it is a 0/0 and dies to
// the state-based actions.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bcea5e7e-655d-48ea-871e-28ca49e5c574",
		Name:            "Drift of the Dead",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{{
			Label:     "Drift of the Dead's power and toughness are each equal to the number of snow lands you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := len(permanentsControlledByMatching(g, source.Controller, And(Land(), Snow())))
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
