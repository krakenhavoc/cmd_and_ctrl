package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enhanced Awareness — Instant {4}{U} (Fate Reforged,
// #5552):
//
//	"Draw three cards, then discard a card."
//
// Ancestral Reminiscence's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "397964db-465a-407d-9dc7-d0a4944473ac",
		Name:         "Enhanced Awareness",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
