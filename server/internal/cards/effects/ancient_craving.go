package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Craving — Sorcery {3}{B} (Reprint, #5908):
//
//	"You draw three cards and you lose 3 life."
//
// Night's Whisper's shape at three: the draw, then the life loss once any
// prompt the draw asks has been answered (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78725353-9274-420a-b722-add0f43c444e",
		Name:         "Ancient Craving",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			GainLife{Amount: -3},
		),
	})
}
