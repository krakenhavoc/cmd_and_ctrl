package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catalog — Instant {2}{U} (Urza's Saga, #5624):
//
//	"Draw two cards, then discard a card."
//
// Three Steps Ahead's loot (drawTwoThenDiscardOne): both cards are drawn
// first, so either may be the discard, which the caster picks from their
// own hand (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "965139fb-2e58-4ef2-b639-3ceaa2995c00",
		Name:         "Catalog",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawTwoThenDiscardOne(item, ctx, 0)
		},
	})
}
