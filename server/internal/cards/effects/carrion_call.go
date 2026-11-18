package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion Call — Instant {3}{G}:
//
//	"Create two 1/1 green Phyrexian Insect creature tokens with infect."
//
// Infect is the engine's damage tail (ADR 0056, #748), granted to the tokens
// through the template's printed keywords, so it is real on a token the moment
// it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f1e8f25-ee23-40a6-a3f2-369bb960267f",
		Name:         "Carrion Call",
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 green Phyrexian Insect with infect"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
