package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Join the Ranks — Instant {3}{W} (Worldwake, #5582):
//
//	"Create two 1/1 white Soldier Ally creature tokens."
//
// Raise the Alarm's shape with the Soldier Ally token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa16307a-9c1f-4538-872c-95206e2b7a6e",
		Name:         "Join the Ranks",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 white Soldier Ally"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
