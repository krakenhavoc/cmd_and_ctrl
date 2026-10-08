package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scatter to the Winds — Instant {1}{U}{U}:
//
//	"Counter target spell.
//	 Awaken 3—{4}{U}{U}"
//
// ADR 0135 §3 (#2411): the counter, then the awaken land (CR 702.113a),
// each target checked on its own (CR 608.2b): a spell that has already
// left the stack still lets the land awaken.
//
// No simplifications.
func init() {
	t := TargetSpell("target spell")
	Register(Spec{
		OracleID:     "13d600e6-86b4-4813-8f34-75f23717f433",
		Name:         "Scatter to the Winds",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(3, "{4}{U}{U}", t),
		},
		OnResolve: AwakenAfter(3, func(_ *game.StackItem, ctx *Context) error {
			if s, ok := ctx.ClauseTarget(0); ok {
				return CounterTarget{StackID: s.ID}.Apply(ctx)
			}
			return nil
		}),
	})
}
