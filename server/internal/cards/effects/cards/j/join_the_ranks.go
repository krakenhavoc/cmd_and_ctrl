package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Join the Ranks — Instant {3}{W} (Worldwake, #5582):
//
//	"Create two 1/1 white Soldier Ally creature tokens."
//
// Raise the Alarm's shape with the Soldier Ally token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa16307a-9c1f-4538-872c-95206e2b7a6e",
		Name:         "Join the Ranks",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.TokenCard("1/1 white Soldier Ally"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
