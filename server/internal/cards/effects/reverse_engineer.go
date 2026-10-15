package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reverse Engineer — Sorcery {3}{U}{U}: "Improvise ... Draw three cards."
//
// Improvise is Spec.TapCost; the draw is Divination's. No simplification.
func init() {
	Register(Spec{
		OracleID:     "f56596be-e12f-4ab6-af61-3ef8dc87134c",
		Name:         "Reverse Engineer",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
