package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Touch of Brilliance — Sorcery {3}{U} (Portal, #5630):
//
//	"Draw two cards."
//
// One draw step for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6365aba1-78d3-416c-89cd-9449578eedbf",
		Name:         "Touch of Brilliance",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    ResolveSteps(DrawCards{N: 2}),
	})
}
