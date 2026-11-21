package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorpion Sentinel — Artifact Creature — Robot Scorpion {1}{U}, 1/4
// (Final Fantasy, #4462):
//
//	"As long as you control seven or more lands, this creature gets
//	 +3/+0."
//
// Woodborn Behemoth's shape: a self-only layer 7c modification whose
// condition is read on every layer pass, so the bonus comes and goes with
// the seventh land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e53fe9f-9995-47ff-85e9-f09b70d7af15",
		Name:         "Scorpion Sentinel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as you control seven or more lands, this creature gets +3/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.B10LandsControlled(g, source.Controller) >= 7
			},
			Apply: effects.PumpBy(3, 0),
		}},
	})
}
