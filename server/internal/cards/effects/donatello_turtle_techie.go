package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello, Turtle Techie — {3}{U} Legendary Creature — Mutant Ninja Turtle
// 3/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Donatello enters, if you control an artifact, draw a card."
//
// An intervening "if" (CR 603.4): the trigger needs an artifact you control
// as Donatello enters, and again as it resolves, or it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f84850bc-6348-449e-bd82-bb39e2119bec",
		Name:         "Donatello, Turtle Techie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *Context) bool { return b03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && b03ArtifactsControlled(g, source.Controller) > 0
			}, "Donatello, Turtle Techie — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
