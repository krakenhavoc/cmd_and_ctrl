package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Attendant — Artifact Creature — Homunculus {2}{U}{U}, */3:
//
//	"Flying
//	 Filigree Attendant's power is equal to the number of artifacts you
//	 control."
//
// A layer 7a characteristic-defining ability on power only (Master of
// Etherium's count, b15ArtifactsControlled, which reads post-layer types
// and counts the Attendant itself).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab7c9e72-8f30-47f7-918c-cd3c0593d704",
		Name:            "Filigree Attendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of artifacts you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = b15ArtifactsControlled(g, source.Controller)
			},
		}},
	})
}
