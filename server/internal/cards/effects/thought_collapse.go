package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Collapse — Instant {1}{U}{U} (Jumpstart):
//
//	"Counter target spell. Its controller mills three cards."
//
// Didn't Say Please under another name, sharing its body: the spell's
// controller is read before the counter moves it off the stack, a
// library holding fewer than three mills what it has, and milling out is
// not losing (CR 701.17b). A target gone in response fizzles the whole
// spell, so nobody mills.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a265af8-448d-402b-b728-b7ed64521842",
		Name:         "Thought Collapse",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b32CounterTargetThenControllerMills(ctx, 3)
		},
	})
}
