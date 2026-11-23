package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Birthday Escape — Sorcery {U}:
//
//	"Draw a card. The Ring tempts you."
//
// In printed order (CR 608.2c): a draw can pause (a dredge offer), and the
// tempt waits for the answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f958678-2a32-4a30-8fe7-47df9ada6b2f",
		Name:         "Birthday Escape",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		// The Ring tempts after the draw's prompt is answered (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 1}, effects.TheRingTemptsYou{}),
	})
}
