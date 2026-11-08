package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biogenic Upgrade — Sorcery {4}{G}{G} (Foundations, #4465):
//
//	"Distribute three +1/+1 counters among one, two, or three target
//	 creatures, then double the number of +1/+1 counters on each of those
//	 creatures."
//
// The distribution is Synchronized Charge's, announced with the targets
// (CR 601.2d), each target getting at least one. The doubling is Bristly
// Bill's (doublePlusOneCountersOnEach) over the targets still legal as it
// resolves, and it runs after the counters land, so it counts them and any
// the creature already had: a bare creature given one counter ends with
// two. A target gone by then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30812870-87c0-469c-95a5-811d5a181e72",
		Name:         "Biogenic Upgrade",
		Completeness: CompletenessFull,
		Targets: TargetCreature("one, two, or three target creatures").
			WithCount(1, 3).Dividing(Divide(3)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return PutDividedCounters(ctx, game.CounterPlusOne) }),
			Instruction(func(ctx *Context) error {
				those := map[uuid.UUID]bool{}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						those[t.ID] = true
					}
				}
				return doublePlusOneCountersOnEach(ctx.Game, ctx.Item, func(c game.Card) bool { return those[c.InstanceID] })
			}),
		),
	})
}
