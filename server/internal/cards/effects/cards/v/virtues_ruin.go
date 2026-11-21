package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtue's Ruin — Sorcery {2}{B} (Portal, #5630):
//
//	"Destroy all white creatures."
//
// Mass Calcify's sweep, turned round: only white creatures die.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc070e97-73ac-4028-9432-ff2012f2e778",
		Name:         "Virtue's Ruin",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature that isn't white stays.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.OfColor("W"))}.Apply(ctx)
		},
	})
}
