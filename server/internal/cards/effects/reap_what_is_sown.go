package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reap What Is Sown — Instant {1}{G}{W} (Born of the Gods,
// #5559):
//
//	"Put a +1/+1 counter on each of up to three target creatures."
//
// Leo's Guidance's first half (putPlusOneCounterOnEachLegalTarget):
// one counter on each pick still legal as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd6cec11-5c66-4f0e-a12f-9b483a09a777",
		Name:         "Reap What Is Sown",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return putPlusOneCounterOnEachLegalTarget(ctx.Game, item)
		},
	})
}
