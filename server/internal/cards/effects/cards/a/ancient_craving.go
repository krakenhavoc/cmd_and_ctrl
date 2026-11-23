package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Craving — Sorcery {3}{B} (Reprint, #5908):
//
//	"You draw three cards and you lose 3 life."
//
// Night's Whisper's shape at three: the draw, then the life loss once any
// prompt the draw asks has been answered (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78725353-9274-420a-b722-add0f43c444e",
		Name:         "Ancient Craving",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.GainLife{Amount: -3},
		),
	})
}
