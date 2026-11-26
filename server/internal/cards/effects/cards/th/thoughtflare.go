package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtflare — Instant {3}{U}{R} (Return to Ravnica, S82 OLD):
//
//	"Draw four cards, then discard two cards."
//
// Draw first, so a drawn card is a legal discard, then the controller
// chooses two cards to discard (DrawThenDiscard, as Careful Study's). A hand
// of fewer than two discards what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faad0f02-da5b-47f4-9634-0ea6ee2ff4f6",
		Name:         "Thoughtflare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 4, Discards: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 4, 2, "")
		},
	})
}
