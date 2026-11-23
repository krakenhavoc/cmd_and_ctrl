package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Fires — Sorcery {1}{R}:
//
//	"Underworld Fires deals 1 damage to each creature and each
//	 planeswalker. If a permanent dealt damage this way would die this
//	 turn, exile it instead."
//
// "A permanent dealt damage this way" (ADR 0108 §1 decision 3): each
// creature and planeswalker that was dealt more than 0 damage is
// marked, so a planeswalker that loses its last loyalty is exiled.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed3d113d-f2c3-4ef3-8bdd-ed4824a22ab0",
		Name:         "Underworld Fires",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachExileIfDealtDies(ctx, effects.Or(effects.Creature(), effects.Planeswalker()), 1, true)
		},
	})
}
