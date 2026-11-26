package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spontaneous Combustion — Instant {1}{B}{R} (Conspiracy, S82 OLD batch 364,
// #5557):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Spontaneous Combustion deals 3 damage to each creature."
//
// The sacrifice is an additional cost (CR 118.8), paid as the spell is cast;
// the sacrificed creature is gone before the damage, so it takes none. The
// damage is Rolling Temblor's sweep, dealt to every creature at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0f81064d-01cd-4a61-8aea-2544e4d14b70",
		Name:           "Spontaneous Combustion",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 3)
		},
	})
}
