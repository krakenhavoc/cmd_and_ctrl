package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curate — Instant {1}{U}:
//
//	"Surveil 2.
//	 Draw a card."
//
// Consider's shape with a deeper surveil: the draw rides the surveil's
// keyed tail, so it happens after the player has answered the prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75e8907d-f726-4a63-ba8b-c0fd4936734e",
		Name:         "Curate",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			c := ctx.Controller()
			return Surveil{Player: c, N: 2, ThenTail: game.KeyedTail{Ref: thenDrawACard, Player: c}}.Apply(ctx)
		},
	})
}
