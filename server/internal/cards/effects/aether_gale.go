package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Gale — Sorcery {3}{U}{U} (EDHREC rank 3675):
//
//	"Return six target nonland permanents to their owners' hands."
//
// Six targets, exactly six, any player's, tokens included — so it
// is uncastable with fewer than six nonland permanents on the table,
// exactly as in paper (CR 601.2c). Ashes to Ashes's multi-target
// shape at count six: every announced permanent still legal at
// resolution goes back to its owner's hand through the simultaneous
// bounce path, and one that left in response is skipped rather than
// fizzling the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b21d6482-1b3a-47e6-98a5-3067f5f3818b",
		Name:         "Aether Gale",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("six target nonland permanents", Nonland()).WithCount(6, 6),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b35BounceChosenPermanents(ctx)
		},
	})
}
