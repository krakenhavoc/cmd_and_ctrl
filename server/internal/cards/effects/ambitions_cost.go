package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambition's Cost — Sorcery {3}{B}:
//
//	"You draw three cards and you lose 3 life."
//
// Night's Whisper scaled up. Same DrawCards + ChangePlayerLife
// composition aimed at the controller.
func init() {
	Register(Spec{
		OracleID:     "84de4fec-2f38-4293-93d3-b3882c5aac14",
		Name:         "Ambition's Cost",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: CompletenessFull,
		// The life loss waits for a prompt the draws ask (CR 608.2c, #4155).
		OnResolve: ResolveSteps(DrawCards{N: 3}, GainLife{Amount: -3}),
	})
}
