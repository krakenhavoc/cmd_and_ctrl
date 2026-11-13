package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miscalculation — Instant {1}{U} (Urza's Legacy, #5623):
//
//	"Counter target spell unless its controller pays {2}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Convolute's soft counter with a {2} tax (CounterUnlessPaid), and the
// shared Cycling row (CR 702.29a). The targeted spell's controller is asked
// at resolution; a "pay" they can't fund is a decline. A target that has
// already left the stack is a fizzle (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9078db26-3a54-4a11-9aae-e45c38da7226",
		Name:         "Miscalculation",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     "{2}",
				Question: "Miscalculation — pay {2} or your spell is countered",
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
