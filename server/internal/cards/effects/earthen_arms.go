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
			// Whether awaken was paid rides the tail as its N: the cost the
			// spell was cast with is not on the asking item it rebuilds.
			then := ctx.ItemTail(earthenArmsAwaken)
			if ctx.PaidAltCost(awakenKey) {
				then.N = 1
			}
			return ctx.Game.AddCounterByKeyedForEffect(item.Controller, id, game.CounterPlusOne, 2, then)
		},
	})
}

// earthenArmsAwaken is the awaken, once the first two counters have landed; the
// tail's N is 1 when awaken was paid.
// An on-disk key: never renamed, never reused.
var earthenArmsAwaken = CardTail("earthen-arms/awaken", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N == 0 {
		return nil
	}
	return awakenPaidClause(ctx, 4)
})
