package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx's Revelation — Instant {X}{W}{U}{U} (EDHREC rank 3379):
//
//	"You gain X life and draw X cards."
//
// The Azorius control finisher. X is the announced value the cost
// engine charged; the life is gained before the cards are drawn, in
// printed order, so a lifegain payoff goes on the stack above the
// draws' payoffs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71d83fca-e40e-4d0e-956d-d0d6da9cc472",
		Name:         "Sphinx's Revelation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b32GainLifeThenDraw(item, ctx, ctx.X())
		},
	})
}
