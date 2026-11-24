package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prying Eyes — Instant {4}{U}{U} (Commander Legends):
//
//	"Draw four cards, then discard two cards."
//
// Careful Study's body with four draws and two discards: the discard prompt
// opens after every draw has resolved, so the cards just drawn are legal
// discards, and a hand smaller than two is discarded whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06b60766-a1fb-4ed0-93f6-8423008f75b4",
		Name:         "Prying Eyes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 4, 2, "")
		},
	})
}
