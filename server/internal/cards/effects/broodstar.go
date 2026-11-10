package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodstar — Creature — Beast {8}{U}{U}, */* (Foundations Commander):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying
//	 Broodstar's power and toughness are each equal to the number of
//	 artifacts you control."
//
// Affinity is the shared self cost modifier (CR 702.41a, generic mana only).
// The size is a layer 7a characteristic-defining ability (CR 604.3, 613.4a)
// over the artifacts its controller controls, counted every recompute. It
// is not an artifact, so it never counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "428cc1cb-a491-4cc4-9436-28ee8acc37d6",
		Name:            "Broodstar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Static: []game.StaticAbility{{
			Label:     "Broodstar's power and toughness are each equal to the number of artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := stickerArtifactsControlledBy(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
