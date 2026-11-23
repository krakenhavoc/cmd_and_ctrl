package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saproling Migration — Sorcery {1}{G} (Foundations Jumpstart):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Create two 1/1 green Saproling creature tokens. If this spell was
//	 kicked, create four of those tokens instead."
//
// Burst Lightning's kicker read (ctx.WasKicked, the record the payment
// left on the stack item) picks two tokens or four.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "bbbb80bd-76fa-4155-929b-4a3565d1cb35",
		Name:          "Saproling Migration",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 2, CreatureTokens: 2},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 4
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
		},
	})
}
