package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Countermand — Instant {2}{U}{U} (Journey into Nyx, #5558):
//
//	"Counter target spell. Its controller mills four cards."
//
// Didn't Say Please with four cards milled: the countered spell's
// controller mills.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43402a6c-9fba-4a80-90b3-14bac05c574b",
		Name:         "Countermand",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b32CounterTargetThenControllerMills(ctx, 4)
		},
	})
}
