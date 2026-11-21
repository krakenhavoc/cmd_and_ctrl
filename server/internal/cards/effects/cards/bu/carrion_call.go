package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8f1e8f25-ee23-40a6-a3f2-369bb960267f",
		Name:         "Carrion Call",
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.TokenCard("1/1 green Phyrexian Insect with infect"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
