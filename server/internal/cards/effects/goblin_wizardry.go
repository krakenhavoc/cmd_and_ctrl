package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Wizardry — Instant {3}{R} (Ravnica: Clue Edition, #5929):
//
//	"Create two 1/1 red Goblin Wizard creature tokens with prowess."
//
// Two tokens from the "1/1 red Goblin Wizard with prowess" row; prowess is
// the engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f621dcaa-c500-4a17-a65a-60d9b5516a3b",
		Name:         "Goblin Wizardry",
		Purpose:      game.Purpose{Tokens: 2},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 red Goblin Wizard with prowess"), N: 2}.Apply(ctx)
		},
	})
}
