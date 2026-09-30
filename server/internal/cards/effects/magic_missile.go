package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magic Missile — Sorcery {1}{R}{R} (#1658, unblocked by #1656):
//
//	"This spell can't be countered.
//	 Magic Missile deals 3 damage divided as you choose among one,
//	 two, or three targets."
//
// Arc Lightning's division with the Spec.CantBeCountered (S23) rider —
// an unconditional flag read at the counter choke point, no cast-time
// dependency the way Banefire's "if X is 5 or more" would need.
func init() {
	Register(Spec{
		OracleID:        "e48b8d78-3f8b-4bf8-853a-76c435bc31f0",
		Name:            "Magic Missile",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetAny().WithCount(1, 3).Dividing(Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
