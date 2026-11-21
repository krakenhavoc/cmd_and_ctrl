package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Wizardry — Instant {3}{R} (Ravnica: Clue Edition, #5929):
//
//	"Create two 1/1 red Goblin Wizard creature tokens with prowess."
//
// Two tokens from the "1/1 red Goblin Wizard with prowess" row; prowess is
// the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f621dcaa-c500-4a17-a65a-60d9b5516a3b",
		Name:         "Goblin Wizardry",
		Purpose:      game.Purpose{Tokens: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin Wizard with prowess"), N: 2}.Apply(ctx)
		},
	})
}
