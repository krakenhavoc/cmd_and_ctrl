package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Didn't Say Please — Instant {1}{U}{U} (EDHREC rank 3446):
//
//	"Counter target spell. Its controller mills three cards."
//
// Cancel with a mill stapled on. The spell's controller is read
// before the counter moves it off the stack; a library holding fewer
// than three mills what it has, and milling out is not losing the
// game (CR 701.17b).
// A target gone in response fizzles the whole spell — no mill, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b90bc464-a95a-42e4-9d9a-4b0882eb57ba",
		Name:         "Didn't Say Please",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b32CounterTargetThenControllerMills(ctx, 3)
		},
	})
}
