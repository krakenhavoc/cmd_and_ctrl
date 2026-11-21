package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyroclasm — "Pyroclasm deals 2 damage to each creature."
//
// Damage, not destruction, and the distinction has consequences: an
// X/3 lives, damage prevention applies, and nothing dies until the
// lethal-damage state-based action runs at the next boundary. Since
// S23 that SBA sweep destroys its whole doomed set as one
// simultaneous event, so a Blood Artist that Pyroclasm killed sees
// every other creature it killed (see game/simultaneous.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c",
		Name:         "Pyroclasm",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 2)
		},
	})
}
