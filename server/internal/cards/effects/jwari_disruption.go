package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jwari Disruption // Jwari Ruins — Modal DFC. Front: Instant {1}{U}:
// "Counter target spell unless its controller pays {1}." Back: Land
// (mdfc_lands.go).
//
// The same soft counter as Daze, without the alternative cost: the
// spell's controller is asked at resolution and pays or loses the spell.
// A target that has already left the stack is a fizzle (CR 608.2b).
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "941a4b14-ea2a-4bd0-8cc2-d609f80df32c",
		Name:         "Jwari Disruption",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Jwari Disruption — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
