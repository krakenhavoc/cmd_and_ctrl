package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0753aee4-33db-48c5-9854-16a9d91535b2",
		Name:         "Scholar of Stars",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *Context) bool { return b03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && b03ArtifactsControlled(g, source.Controller) > 0
			}, "Scholar of Stars — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
