package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaelicat — Creature — Cat {2}{W}, 1/3 (Final Fantasy, #4462):
//
//	"Flying, vigilance
//	 As long as you control two or more artifacts, this creature gets
//	 +2/+0."
//
// Scorpion Sentinel's shape keyed on artifacts: a self-only layer 7c
// modification (CR 613.4c) whose condition counts the artifacts its
// controller controls on the post-layer type line every pass, so an
// animated Vehicle or a Treasure counts and the bonus leaves with the
// second artifact. Flying and vigilance ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34bdeca4-b1e0-4fa3-afe4-bec743a080b8",
		Name:            "Gaelicat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:    "As long as you control two or more artifacts, this creature gets +2/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.B15ArtifactsControlled(g, source.Controller) >= 2
			},
			Apply: effects.PumpBy(2, 0),
		}},
	})
}
