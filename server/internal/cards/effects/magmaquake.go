package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmaquake — Instant {X}{R}{R}:
//
//	"Magmaquake deals X damage to each creature without flying and each
//	 planeswalker."
//
// Nothing is targeted. Creatures are read through the layered view, so a
// granted flying spares a creature; every planeswalker takes X, whoever
// controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a52037df-0be1-4c99-b9eb-b033157af203",
		Name:         "Magmaquake",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Or(And(Creature(), WithoutKeyword("flying")), Planeswalker()), ctx.X())
		},
	})
}
