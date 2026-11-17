package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gerrard's Wisdom — Sorcery {2}{W}{W} (Weatherlight, #5629):
//
//	"You gain 2 life for each card in your hand."
//
// The hand is counted once, as the spell resolves (CR 608.2h); the
// spell itself is on the stack then, not in the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e30e8f2-f437-4620-a3bc-dd9c29ae6570",
		Name:         "Gerrard's Wisdom",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Amount: 2 * b14HandSize(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
