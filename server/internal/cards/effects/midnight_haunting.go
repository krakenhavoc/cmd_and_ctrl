package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Haunting — Instant {2}{W}:
//
//	"Create two 1/1 white Spirit creature tokens with flying."
//
// The Spirit token is the one Lingering Souls makes; flying is real on it as
// it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a79e21fc-89bd-4547-8bef-5510554e1431",
		Name:         "Midnight Haunting",
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 white Spirit with flying"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
