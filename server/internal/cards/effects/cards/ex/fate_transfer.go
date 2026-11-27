package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fate Transfer — Instant {1}{U/B} (Shadowmoor, S82 older sets):
//
//	"Move all counters from target creature onto another target creature."
//
// Two clauses, the second distinct from the first (Incremental Growth's shape).
// Every kind of counter on the first creature moves, kind by kind, as The
// Ozolith's does. If either target is illegal as the spell resolves no counter
// is moved: the first creature has nothing to give up or the second cannot take
// it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a07451f-70b3-4bfd-af54-6b4182369652",
		Name:         "Fate Transfer",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature").Then(
			effects.Distinct(effects.TargetCreature("another target creature"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) < 2 || !ctx.IsTargetLegal(item.Targets[0]) || !ctx.IsTargetLegal(item.Targets[1]) {
				return nil
			}
			from, to := item.Targets[0].ID, item.Targets[1].ID
			src, ok := ctx.Game.LookupCardForEffect(from)
			if !ok {
				return nil
			}
			counts := make(map[string]int, len(src.Counters))
			for kind, n := range src.Counters {
				counts[kind] = n
			}
			for kind, n := range counts {
				if n <= 0 {
					continue
				}
				if err := ctx.Game.AddCounterForEffect(from, kind, -n); err != nil {
					return err
				}
				if err := ctx.Game.AddCounterForEffect(to, kind, n); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
