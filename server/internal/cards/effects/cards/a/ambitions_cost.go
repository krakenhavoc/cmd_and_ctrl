package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambition's Cost — Sorcery {3}{B}:
//
//	"You draw three cards and you lose 3 life."
//
// Night's Whisper scaled up. Same DrawCards + ChangePlayerLife
// composition aimed at the controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84de4fec-2f38-4293-93d3-b3882c5aac14",
		Name:         "Ambition's Cost",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: effects.CompletenessFull,
		// The life loss waits for a prompt the draws ask (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 3}, effects.GainLife{Amount: -3}),
	})
}
