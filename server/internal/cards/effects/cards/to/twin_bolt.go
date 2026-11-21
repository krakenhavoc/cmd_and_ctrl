package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twin Bolt — Instant {1}{R} (#1658, unblocked by #1656):
//
//	"Twin Bolt deals 2 damage divided as you choose among one or two
//	 targets."
//
// Forked Bolt at instant speed for one more mana. Same shape, same
// division rule (CR 601.2d): each of the one or two targets gets at
// least 1, the shares summing to 2.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "970dc070-3668-4584-9904-2897b27bb806",
		Name:         "Twin Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(1, 2).Dividing(effects.Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
