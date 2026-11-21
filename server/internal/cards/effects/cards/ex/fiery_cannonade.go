package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Cannonade — Instant {2}{R} (Ixalan, #5532):
//
//	"Fiery Cannonade deals 2 damage to each non-Pirate creature."
//
// Breath Weapon's sweep with Pirate for Dragon: every creature that
// isn't a Pirate as it resolves is dealt 2 damage at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e108b31-2ea3-4129-b324-a2d31c78bbe6",
		Name:         "Fiery Cannonade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Pirate"))), 2)
		},
	})
}
