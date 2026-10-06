package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ed3d113d-f2c3-4ef3-8bdd-ed4824a22ab0",
		Name:         "Underworld Fires",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachExileIfDealtDies(ctx, Or(Creature(), Planeswalker()), 1, true)
		},
	})
}
