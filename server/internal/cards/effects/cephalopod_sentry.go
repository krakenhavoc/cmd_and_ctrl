package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cephalopod Sentry — Artifact Creature — Phyrexian Squid {2}{W}{U}, */5
// (Phyrexia: All Will Be One):
//
//	"Flying
//	 Cephalopod Sentry's power is equal to the number of artifacts you
//	 control."
//
// The power is a layer 7a characteristic-defining ability (CR 604.3,
// 613.4a) counted on every recompute, Broodstar's shape. The Sentry is an
// artifact, so it counts itself. Toughness is the printed 5.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8270c29-1887-4d3f-b70f-6b79ada8fe47",
		Name:            "Cephalopod Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Cephalopod Sentry's power is equal to the number of artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = stickerArtifactsControlledBy(g, source.Controller)
			},
		}},
	})
}
