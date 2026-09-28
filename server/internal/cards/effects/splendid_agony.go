package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splendid Agony — Instant {2}{B} (#1658, unblocked by #1656):
//
//	"Distribute two -1/-1 counters among one or two target
//	 creatures."
//
// The current Oracle wording (checked against the Sep-23 dump, all
// six printings agree) distributes fresh -1/-1 counters — it does not
// move counters from one creature to another, so it needs nothing
// beyond PutDividedCounters with CounterMinusOne: one or two targets,
// each getting at least 1 of the 2 (CR 601.2d).
func init() {
	Register(Spec{
		OracleID:     "e9958344-13d5-4f7b-acc8-904e5d99b90e",
		Name:         "Splendid Agony",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutDividedCounters(ctx, game.CounterMinusOne)
		},
	})
}
