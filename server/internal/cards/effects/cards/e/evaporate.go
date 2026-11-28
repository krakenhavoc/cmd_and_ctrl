package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evaporate — Sorcery {2}{R} (Homelands, S82 OLD b824, #5635):
//
//	"Evaporate deals 1 damage to each white and/or blue creature."
//
// One damage instance to every creature that is white, blue or both, read
// through the layers (CR 105.2), mine included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f9e3b79-6ecb-4741-a9f4-fd7d3545d5cf",
		Name:         "Evaporate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Or(effects.OfColor("W"), effects.OfColor("U"))), 1)
		},
	})
}
