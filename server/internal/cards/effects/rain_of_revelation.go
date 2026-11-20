package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rain of Revelation — Instant {3}{U} (Modern Horizons, #5519):
//
//	"Draw three cards, then discard a card."
//
// Sift at instant speed: drawThenDiscard 3 and 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d5dfa6b-f685-4925-b93a-ae5815a3c0a7",
		Name:         "Rain of Revelation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
