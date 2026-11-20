package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Big Score — Instant for {3}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create two Treasure tokens."
//
// Thrill of Possibility with a rider. Four permanents' worth of
// artifact triggers in a deck that cares (Reckless Fireweaver,
// Ingenious Artillerist) — two Treasures entering, plus whatever
// the discard itself sets off.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a5cbd257-c836-493e-bb1a-76242619dea2",
		Name:           "Big Score",
		Purpose:        game.Purpose{Draws: 2, Tokens: 2, Answers: game.AnswerValue}, // #4734: draw and Treasure tokens answer nothing
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawAndTreasure(item, ctx, 2, 2)
		},
	})
}
