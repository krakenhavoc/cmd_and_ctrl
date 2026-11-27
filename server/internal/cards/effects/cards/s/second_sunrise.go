package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Second Sunrise — Instant {1}{W}{W} (Mirrodin, #5606):
//
//	"Each player returns to the battlefield all artifact, creature,
//	 enchantment, and land cards in their graveyard that were put there from
//	 the battlefield this turn."
//
// Cry of the Carnarium's "put there from the battlefield this turn" record
// (the card's newest arrival in the graveyard, CR 400.7), over four card types
// and every graveyard. Each card returns under its owner's control, all at
// once as one entry (Pyrrhic Revival's batch door); the set is read before the
// first one moves, and a token is not a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15cef63b-903e-4f1b-a128-e59157e07c36",
		Name:         "Second Sunrise",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB607SecondSunrise(ctx.Game, item)
		},
	})
}
