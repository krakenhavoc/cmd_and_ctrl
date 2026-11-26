package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Shudder — Instant {1}{R} (Zendikar, S82 OLD batch 492, #5583):
//
//	"Seismic Shudder deals 1 damage to each creature without flying."
//
// Rolling Temblor's shape with one damage: every creature without flying
// (read through the layers as it resolves) is dealt 1 damage, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10c0f705-4c23-4761-afa9-4a29cf90e8ae",
		Name:         "Seismic Shudder",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature with flying is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
		},
	})
}
