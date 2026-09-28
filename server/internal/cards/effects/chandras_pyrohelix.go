package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra's Pyrohelix — Instant {1}{R} (#1658, unblocked by #1656):
//
//	"Chandra's Pyrohelix deals 2 damage divided as you choose among
//	 one or two targets."
//
// Forked Bolt / Twin Bolt's shape again: fixed 2, one or two targets,
// each getting at least 1 (CR 601.2d).
func init() {
	Register(Spec{
		OracleID:     "69a37af8-6bcd-42b3-b788-0a357e742cc0",
		Name:         "Chandra's Pyrohelix",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
