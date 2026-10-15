package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metallic Rebuke — Instant {2}{U}: "Improvise ... Counter target spell
// unless its controller pays {3}."
//
// Improvise is Spec.TapCost; the counter is Dazzling Denial's
// CounterUnlessPaid. No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4920e95-a6fc-4373-a0d6-244f51851364",
		Name:         "Metallic Rebuke",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			if ctx.Game.StackItemForEffect(stackID) == nil {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  stackID,
				Cost:     "{3}",
				Question: "Metallic Rebuke — pay {3} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
