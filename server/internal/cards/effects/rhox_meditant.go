package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "832c6834-2cc7-41d5-b881-37c24c13ab8a",
		Name:         "Rhox Meditant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control a green permanent",
				Test: func(ctx *Context) bool {
					return b18ControlsPermanentOfColor(ctx.Game, ctx.Controller(), ctx.TextColor("G"))
				},
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && b18ControlsPermanentOfColor(g, source.Controller, source.TextColor("G"))
			}, "Rhox Meditant — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
