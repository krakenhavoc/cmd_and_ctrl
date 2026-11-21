package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yamabushi's Storm — Sorcery {1}{R}:
//
//	"Yamabushi's Storm deals 1 damage to each creature. If a creature
//	 dealt damage this way would die this turn, exile it instead."
//
// Each creature is told what it was dealt (ADR 0108 §1 decision 4): one
// whose damage was prevented is not marked.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dfeb0c5-85d6-48fb-b924-d7b77f4b89d6",
		Name:         "Yamabushi's Storm",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachExileIfDealtDies(ctx, effects.Creature(), 1, false)
		},
	})
}
