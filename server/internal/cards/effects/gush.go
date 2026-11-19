package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gush — Instant {4}{U} (Mercadian Masques, #5837):
//
//	"You may return two Islands you control to their owner's hand rather
//	 than pay this spell's mana cost.
//	 Draw two cards."
//
// Daze's alternative cost with a count (ReturnInstead over a clause of
// exactly two, #5837): the caster names two Islands they control, and both
// go back to their owners' hands as a cost with Gush already on the stack
// (CR 118.9, 601.2h, 400.3). "Islands" is the land type (CR 205.3i), so a
// nonbasic Island counts. With one Island the offer can't be paid (CR
// 118.3) and isn't shown.
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
