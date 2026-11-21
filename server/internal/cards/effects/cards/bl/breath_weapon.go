package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breath Weapon — Instant {2}{R}:
//
//	"Breath Weapon deals 2 damage to each non-Dragon creature."
//
// Damage, not destruction: indestructible and prevention work as
// printed, and the deaths are the state-based sweep's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3dd1f6d-f7d7-4358-8139-7495404f29c7",
		Name:         "Breath Weapon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Not(effects.CmdrB5Dragon())), 2)
		},
	})
}
