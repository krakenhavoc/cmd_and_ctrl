package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anger of the Gods — Sorcery {1}{R}{R}:
//
//	"Anger of the Gods deals 3 damage to each creature. If a creature
//	 dealt damage this way would die this turn, exile it instead."
//
// Each creature is told what it was dealt (game.DealDamageEachEach
// ThenForEffect, ADR 0108 §1 decision 4): one behind a shield that
// prevented all 3 is not marked, and dies normally if it dies later.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3a7fe095-8278-4b1d-bec4-19b35bdcdd1b",
		Name:         "Anger of the Gods",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachExileIfDealtDies(ctx, Creature(), 3, false)
		},
	})
}
