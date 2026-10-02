package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3dfeb0c5-85d6-48fb-b924-d7b77f4b89d6",
		Name:         "Yamabushi's Storm",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachExileIfDealtDies(ctx, Creature(), 1, false)
		},
	})
}
