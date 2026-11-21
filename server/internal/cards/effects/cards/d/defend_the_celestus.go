package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c079a636-cfd4-49cd-bcd4-54375b3fa355",
		Name:         "Defend the Celestus",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("one, two, or three target creatures you control", effects.YouControl()).
			WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
