package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gigantoad — Creature — Frog {3}{G}, 4/4 (Final Fantasy, #4462):
//
//	"As long as you control seven or more lands, this creature gets
//	 +2/+2."
//
// Scorpion Sentinel's shape at +2/+2: a self-only layer 7c modification
// (CR 613.4c) whose condition is read on every layer pass, so the bonus
// comes and goes with the seventh land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa5e4aa5-e2cb-4c8c-91cb-554bec0f21d4",
		Name:         "Gigantoad",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as you control seven or more lands, this creature gets +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.B10LandsControlled(g, source.Controller) >= 7
			},
			Apply: effects.PumpBy(2, 2),
		}},
	})
}
