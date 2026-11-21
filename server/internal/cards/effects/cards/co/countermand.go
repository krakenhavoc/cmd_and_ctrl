package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Countermand — Instant {2}{U}{U} (Journey into Nyx, #5558):
//
//	"Counter target spell. Its controller mills four cards."
//
// Didn't Say Please with four cards milled: the countered spell's
// controller mills.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43402a6c-9fba-4a80-90b3-14bac05c574b",
		Name:         "Countermand",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B32CounterTargetThenControllerMills(ctx, 4)
		},
	})
}
