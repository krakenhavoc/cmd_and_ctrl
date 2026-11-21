package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metallic Rebuke — Instant {2}{U}: "Improvise ... Counter target spell
// unless its controller pays {3}."
//
// Improvise is Spec.TapCost; the counter is Dazzling Denial's
// CounterUnlessPaid. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4920e95-a6fc-4373-a0d6-244f51851364",
		Name:         "Metallic Rebuke",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			if ctx.Game.StackItemForEffect(stackID) == nil {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  stackID,
				Cost:     "{3}",
				Question: "Metallic Rebuke — pay {3} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
