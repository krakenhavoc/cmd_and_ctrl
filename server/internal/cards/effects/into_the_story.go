package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Story — Instant {5}{U}{U}:
//
//	"This spell costs {3} less to cast if an opponent has seven or more
//	 cards in their graveyard.
//	 Draw four cards."
//
// #746: a conditional self cost modifier.
func init() {
	Register(Spec{
		OracleID:     "f290c2e4-ab52-44d4-bdeb-31aeb835b18d",
		Name:         "Into the Story",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if an opponent has seven or more cards in their graveyard.",
				AnOpponentHasCardsInGraveyardAtLeast(7)),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	})
}
