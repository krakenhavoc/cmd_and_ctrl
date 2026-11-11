package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keep Watch — Instant {2}{U}:
//
//	"Draw a card for each attacking creature."
//
// Every attacking creature counts, whoever controls it, counted as the
// spell resolves (CR 608.2h). Outside combat nothing is attacking and the
// spell draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "847b5be5-cb0c-4857-bcd9-c5cfe55fdaeb",
		Name:         "Keep Watch",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: len(AttackingCreatures(ctx.Game))}.Apply(ctx)
		},
	})
}
