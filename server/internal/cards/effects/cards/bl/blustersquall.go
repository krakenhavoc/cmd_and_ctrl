package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blustersquall — Instant {U} (Commander 2015, S82 OLD #5546):
//
//	"Tap target creature you don't control.
//	 Overload {3}{U} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Mizzium Mortars' overload shape: {3}{U} instead of {U}, and the target
// clause deleted, so "target creature you don't control" becomes "each
// creature you don't control". The overloaded spell has no target and taps
// every such creature as it resolves, in one go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dfad8c3-1973-4fc8-971d-f66cacd88070",
		Name:         "Blustersquall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{3}{U}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.PaidAltCost("overload") {
				for _, opp := range ctx.Opponents() {
					if err := effects.B29TapAllCreaturesControlledBy(ctx, opp); err != nil {
						return err
					}
				}
				return nil
			}
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.TapTarget{Target: id}.Apply(ctx)
		},
	})
}
