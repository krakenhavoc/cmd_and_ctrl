package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reviving Dose — Instant {2}{W} (Conspiracy: Take the Crown, S82 older
// sets batch 271, #5540):
//
//	"You gain 3 life.
//	 Draw a card."
//
// Two steps in printed order (CR 608.2c): the life gain, then the draw,
// which waits for any prompt a life-gain replacement asks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad470afc-d1a5-4e64-8192-fb6308ed9bfe",
		Name:         "Reviving Dose",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, LifeGain: 3},
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 3}, effects.DrawCards{N: 1}),
	})
}
