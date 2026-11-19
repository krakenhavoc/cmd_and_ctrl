package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gush — Instant {4}{U}:
//
//	"You may return two Islands you control to their owner's hand rather than
//	 pay this spell's mana cost.
//	 Draw two cards."
//
// Daze's return-instead cost with a clause that names two Islands: both are
// bounced as a cost with the spell already on the stack (CR 601.2h), so a
// countered Gush still costs them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16d8ee99-8ec8-429f-9ba7-818a74b6f910",
		Name:         "Gush",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		AlternativeCosts: []game.AlternativeCost{
			ReturnInstead(
				"Return two Islands you control to their owner's hand",
				PermanentYouControl("two Islands you control", HasSubtype("Island")).WithCount(2, 2),
				"two Islands you control",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
