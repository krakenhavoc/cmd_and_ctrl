package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox Meditant — Creature — Rhino Monk {3}{W}, 2/4 (Conflux, #5587):
//
//	"When this creature enters, if you control a green permanent, draw a
//	 card."
//
// An intervening-if trigger (CR 603.4): with no green permanent as it
// enters there is no trigger, and a green permanent that leaves in
// response makes it draw nothing. "Green" is a word of the source's text
// (CR 612.2), read through ctx.TextColor and source.TextColor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "832c6834-2cc7-41d5-b881-37c24c13ab8a",
		Name:         "Rhox Meditant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control a green permanent",
				Test: func(ctx *effects.Context) bool {
					return effects.B18ControlsPermanentOfColor(ctx.Game, ctx.Controller(), ctx.TextColor("G"))
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B18ControlsPermanentOfColor(g, source.Controller, source.TextColor("G"))
			}, "Rhox Meditant — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
