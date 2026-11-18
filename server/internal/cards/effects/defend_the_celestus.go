package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defend the Celestus — Instant {2}{G}{G} (Innistrad: Midnight Hunt, #5497):
//
//	"Distribute three +1/+1 counters among one, two, or three target
//	 creatures you control."
//
// Armament Dragon's divided counters as a spell: one to three targets, each
// getting at least one counter (CR 601.2d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c079a636-cfd4-49cd-bcd4-54375b3fa355",
		Name:         "Defend the Celestus",
		Completeness: CompletenessFull,
		Targets: TargetCreature("one, two, or three target creatures you control", YouControl()).
			WithCount(1, 3).Dividing(Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
