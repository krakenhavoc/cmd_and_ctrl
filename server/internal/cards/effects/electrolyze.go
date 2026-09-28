package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrolyze — Instant {1}{U}{R} (#1658, unblocked by #1656):
//
//	"Electrolyze deals 2 damage divided as you choose among one or
//	 two targets.
//	 Draw a card."
//
// Forked Bolt's split with card advantage bolted on. The draw is
// unconditional and not part of the division — it happens whether or
// not either target is still legal at resolution (CR 608.2b only
// touches the divided damage).
func init() {
	Register(Spec{
		OracleID:     "07b222d7-24f2-4994-9004-ff6672ebe161",
		Name:         "Electrolyze",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := DealDividedDamage(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
