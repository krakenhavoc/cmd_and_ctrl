package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindstatic — Instant {3}{U}:
//
//	"Counter target spell unless its controller pays {6}."
//
// Convolute's counter at {6}: the spell's controller is asked to pay as it
// resolves, and the target is countered unless they do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "880e0177-f1d2-4bdc-83b4-cae6a0706ef9",
		Name:         "Mindstatic",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     "{6}",
				Question: "Mindstatic — pay {6} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
