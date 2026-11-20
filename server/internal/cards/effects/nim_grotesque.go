package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nim Grotesque — Creature — Zombie {6}{B}, 3/6 (Fifth Dawn, #5604):
//
//	"This creature gets +1/+0 for each artifact you control."
//
// Storm-Kiln Artist's layer 7c count of the artifacts its controller
// controls, read every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed4d4084-739b-43ac-bd2b-a86d1ecd7bbd",
		Name:         "Nim Grotesque",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			Label:     "This creature gets +1/+0 for each artifact you control.",
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactsControlledBy(g, source)
			},
		}},
	})
}
