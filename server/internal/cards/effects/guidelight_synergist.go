package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guidelight Synergist — Artifact Creature — Robot Artificer {3}{W}, 0/4
// (Aetherdrift, #4464):
//
//	"Flying
//	 This creature gets +1/+0 for each artifact you control."
//
// Uthros Research Craft's layer 7c count (artifactsControlledBy) on a
// permanent that is always a creature. The Synergist is an artifact, so
// it counts itself. The count reads each permanent's layered types, so a
// permanent another effect makes an artifact counts too, and it is
// recounted on every layer pass (CR 613.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c0bc3704-9bf9-4039-b3f8-205c1e5eb6e9",
		Name:            "Guidelight Synergist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each artifact you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactsControlledBy(g, source)
			},
		}},
	})
}
