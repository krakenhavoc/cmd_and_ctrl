package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f84850bc-6348-449e-bd82-bb39e2119bec",
		Name:         "Donatello, Turtle Techie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *effects.Context) bool { return effects.B03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B03ArtifactsControlled(g, source.Controller) > 0
			}, "Donatello, Turtle Techie — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
