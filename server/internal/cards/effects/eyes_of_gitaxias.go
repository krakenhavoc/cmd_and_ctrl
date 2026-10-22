package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyes of Gitaxias — Sorcery {2}{U} (March of the Machine):
//
//	"Incubate 3.
//	 Draw a card."
//
// Two instructions in printed order (CR 608.2c): the Incubator
// (CR 701.53a, ADR 0185) first, then the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ef6ca6e-a8b8-4ea6-b13d-a08eb847c12d",
		Name:         "Eyes of Gitaxias",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		// The draw waits for a prompt the Incubator token asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Incubate{N: 3},
			DrawCards{N: 1},
		),
	})
}
