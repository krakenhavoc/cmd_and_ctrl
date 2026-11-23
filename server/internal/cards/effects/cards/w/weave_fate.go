package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weave Fate — Instant {3}{U} (Khans of Tarkir, #5554):
//
//	"Draw two cards."
//
// Touch of Brilliance's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff568a29-e31a-4bc3-b8a3-8fa56f91d52d",
		Name:         "Weave Fate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    effects.ResolveSteps(effects.DrawCards{N: 2}),
	})
}
