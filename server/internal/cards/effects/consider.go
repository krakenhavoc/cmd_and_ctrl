package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consider — Instant {U} (EDHREC rank 405):
//
//	"Surveil 1. (Look at the top card of your library. You may put it
//	 into your graveyard.)
//	 Draw a card."
//
// Surveil queues a prompt and nothing moves until the caster answers
// it, so the draw rides Surveil's Then rather than the next
// statement — the same ordering rule Preordain's Scry-then-draw
// follows: writing the draw as a second primitive would pull a card
// before the surveil decision is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c9bcba6-87b5-4fb3-97ee-6fe5b739337d",
		Name:         "Consider",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			c := ctx.Controller()
			return Surveil{Player: c, N: 1, Then: func(g *game.Game) error {
				return g.DrawNForEffect(c, 1)
			}}.Apply(ctx)
		},
	})
}
