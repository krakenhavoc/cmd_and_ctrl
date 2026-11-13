package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Force Spike — Instant {U} (Legends, #5639):
//
//	"Counter target spell unless its controller pays {1}."
//
// Spell Pierce's pay-unless on any spell, for {1} (CR 118.12): the
// prompt goes to the targeted spell's controller, the spell can't
// resolve while the {1} is outstanding, and a payment they can't fund
// counters it. A target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "896ea218-3019-47e4-b0a0-2487335872bb",
		Name:         "Force Spike",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Force Spike — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
