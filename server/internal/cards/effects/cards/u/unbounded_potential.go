package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unbounded Potential — Instant {1}{W}:
//
//	"Choose one —
//	 • Put a +1/+1 counter on each of up to two target creatures.
//	 • Proliferate. (Choose any number of permanents and/or players,
//	   then give each another counter of each kind already there.)
//	 Entwine {3}{W} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The first bullet may target none, one
// or two creatures and counts only those still legal as it resolves
// (CR 608.2b). Proliferate is the chosen kind (CR 701.34a), asked of
// you. Entwined, the counters go on first and the proliferate after
// (CR 702.42b), so it can add a second +1/+1 counter to a creature the
// first bullet just gave one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d9ae08cf-eef0-407a-ad2c-08b4c89f99f0",
		Name:          "Unbounded Potential",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{3}{W}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Put a +1/+1 counter on each of up to two target creatures.",
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeDoing("Proliferate.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Proliferate{}.Apply(ctx)
				}),
		)),
	})
}
