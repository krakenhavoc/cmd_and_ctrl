package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secure the Wastes — Instant {X}{W} (EDHREC rank 1247):
//
//	"Create X 1/1 white Warrior creature tokens."
//
// An instant-speed army for X. The X is announced with the cast and
// read back with ctx.X(); zero tokens for X=0 is a legal, pointless
// cast, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2347910-d6c6-4681-8316-7ef27056485c",
		Name:         "Secure the Wastes",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker}, // #4734: creature tokens at instant speed
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Warrior"), N: ctx.X()}.Apply(ctx)
		},
	})
}
