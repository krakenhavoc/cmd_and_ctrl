package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrenn's Resolve — Sorcery {1}{R} (EDHREC rank 2131):
//
//	"Exile the top two cards of your library. Until the end of your
//	 next turn, you may play those cards."
//
// Reckless Impulse reprinted under a new name; see that file for the
// shape and the duration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7d8c99c-cd44-4bad-82e6-7c7cff9bd89f",
		Name:         "Wrenn's Resolve",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
		},
	})
}
