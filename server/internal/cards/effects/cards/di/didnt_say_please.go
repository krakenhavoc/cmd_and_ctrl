package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b90bc464-a95a-42e4-9d9a-4b0882eb57ba",
		Name:         "Didn't Say Please",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B32CounterTargetThenControllerMills(ctx, 3)
		},
	})
}
