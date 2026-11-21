package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scholar of Stars — Creature — Human Artificer {3}{U}, 3/2
// (Core Set 2019, #5527):
//
//	"When this creature enters, if you control an artifact, draw a card."
//
// Donatello, Turtle Techie's enters trigger: an intervening if (CR
// 603.4) checked as it triggers and again on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0753aee4-33db-48c5-9854-16a9d91535b2",
		Name:         "Scholar of Stars",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *effects.Context) bool { return effects.B03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B03ArtifactsControlled(g, source.Controller) > 0
			}, "Scholar of Stars — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
