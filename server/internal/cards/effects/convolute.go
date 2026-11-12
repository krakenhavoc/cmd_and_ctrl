package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Convolute — Instant {2}{U} (Ravnica: City of Guilds, #5600):
//
//	"Counter target spell unless its controller pays {4}."
//
// Jwari Disruption's soft counter with a {4} tax (CounterUnlessPaid): the
// targeted spell's controller is asked at resolution and pays or loses the
// spell, and a "pay" they can't fund is a decline. A target that has
// already left the stack is a fizzle (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd201ded-cfaf-4a51-9fcc-6a61b3c910b0",
		Name:         "Convolute",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     "{4}",
				Question: "Convolute — pay {4} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
