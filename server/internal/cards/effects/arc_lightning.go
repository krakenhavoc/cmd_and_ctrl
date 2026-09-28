package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arc Lightning — Sorcery {2}{R} (#1658, unblocked by #1656):
//
//	"Arc Lightning deals 3 damage divided as you choose among one,
//	 two, or three targets."
//
// The fixed-3, one-to-three-targets shape (Inferno Titan's trigger and
// Mogg Mob's activated ability both divide the same 3). Each of the
// chosen targets gets at least 1 of the 3 (CR 601.2d).
func init() {
	Register(Spec{
		OracleID:     "5acc8b39-3c3e-4012-8cfd-ac3c2c4ca982",
		Name:         "Arc Lightning",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 3).Dividing(Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
