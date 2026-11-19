package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "bbbb80bd-76fa-4155-929b-4a3565d1cb35",
		Name:          "Saproling Migration",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tokens: 2, CreatureTokens: 2},
		OptionalCosts: []game.AdditionalCost{Kicker("{4}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 4
			}
			return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
		},
	})
}
