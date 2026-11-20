package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enhanced Awareness — Instant {4}{U} (Fate Reforged,
// #5552):
//
//	"Draw three cards, then discard a card."
//
// Ancestral Reminiscence's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "397964db-465a-407d-9dc7-d0a4944473ac",
		Name:         "Enhanced Awareness",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
