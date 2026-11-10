package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deduce — Instant {1}{U} (Secrets of Strixhaven Commander, #4834):
//
//	"Draw a card. Investigate. (Create a Clue token. It's an artifact
//	 with "{2}, Sacrifice this token: Draw a card.")"
//
// Each sentence waits for the prompt the one before it asks (CR
// 608.2c, #4155), so a replacement choice on the draw settles before the
// Clue is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eeae4161-9f35-45df-8b3c-c55e37f49cd1",
		Name:         "Deduce",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			DrawCards{N: 1},
			Investigate{N: 1},
		),
	})
}
