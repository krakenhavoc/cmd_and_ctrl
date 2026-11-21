package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nourish — Instant {G}{G} (Darksteel, #5605):
//
//	"You gain 6 life."
//
// One GainLife step (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d5e82d3-79ac-4dbe-80e8-1db6dd6c8767",
		Name:         "Nourish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 6},
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 6}),
	})
}
