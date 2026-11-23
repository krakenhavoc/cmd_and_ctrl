package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a52037df-0be1-4c99-b9eb-b033157af203",
		Name:         "Magmaquake",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Or(effects.And(effects.Creature(), effects.WithoutKeyword("flying")), effects.Planeswalker()), ctx.X())
		},
	})
}
