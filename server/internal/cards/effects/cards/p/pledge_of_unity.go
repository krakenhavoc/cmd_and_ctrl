package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pledge of Unity — Instant {1}{G}{W} (War of the Spark):
//
//	"Put a +1/+1 counter on each creature you control. You gain 1 life for each
//	 creature you control."
//
// The counters are the first step and the life the second: the creatures are
// counted as the life is gained (CR 608.2c), after the counters, so a creature
// that left in between is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad501bcb-b838-4716-9759-98c484772285",
		Name:         "Pledge of Unity",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 1)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() {
						n++
					}
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
			}),
		),
	})
}
