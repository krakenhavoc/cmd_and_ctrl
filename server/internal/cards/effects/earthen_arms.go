package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthen Arms — Sorcery {1}{G}:
//
//	"Put two +1/+1 counters on target permanent.
//	 Awaken 4—{6}{G}"
//
// ADR 0135 §3 (#2411): two +1/+1 counters on the target permanent through
// the CR 614 placement window, then the awaken land (CR 702.113a). The two
// clauses may name the same land (CR 601.2c), which then gets six. The
// awaken runs from the first placement's continuation, so a CR 616
// ordering prompt on the first two counters holds the awaken until it is
// answered (CR 608.2c: in the order written).
//
// No simplifications.
func init() {
	t := TargetPermanent("target permanent", Permanent())
	Register(Spec{
		OracleID:     "2146fa14-dc6a-4e13-bed0-dda235784ee6",
		Name:         "Earthen Arms",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(4, "{6}{G}", t),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := awakenSpellCardTarget(ctx)
			if !ok {
				return AwakenIfPaid(ctx, 4)
			}
			return ctx.Game.AddCounterByThenForEffect(item.Controller, id, game.CounterPlusOne, 2,
				func(g *game.Game, _ int) error { return AwakenIfPaid(NewContext(g, item), 4) })
		},
	})
}
