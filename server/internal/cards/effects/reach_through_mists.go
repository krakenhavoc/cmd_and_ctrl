package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reach Through Mists — Instant — Arcane {U} (Champions of Kamigawa,
// #5603):
//
//	"Draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c81ca8ff-92f3-481e-82f7-0673b6c74ea0",
		Name:         "Reach Through Mists",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
