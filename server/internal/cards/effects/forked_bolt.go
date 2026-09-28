package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forked Bolt — Sorcery {R} (#1658, unblocked by #1656):
//
//	"Forked Bolt deals 2 damage divided as you choose among one or
//	 two targets."
//
// One point of removal or two split one-and-one — the fixed-amount,
// bounded-count shape (Inferno Titan's trigger, one printed ability
// lower). The division is announced with the targets (CR 601.2d):
// each target at least 1, the shares summing to 2.
func init() {
	Register(Spec{
		OracleID:     "e0548f0d-4aec-4c39-8c34-52bfa2bf7c60",
		Name:         "Forked Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
