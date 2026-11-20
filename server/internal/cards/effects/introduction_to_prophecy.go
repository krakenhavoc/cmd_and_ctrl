package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Introduction to Prophecy — Sorcery — Lesson {3} (Strixhaven: School of
// Mages, #5502):
//
//	"Scry 2, then draw a card."
//
// Joint Exploration's first two instructions in printed order (CR
// 608.2c): the scry, then the draw, which waits for the scry prompt's
// answer. Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ae1fded-2e77-498a-98e7-4a4908447175",
		Name:         "Introduction to Prophecy",
		Purpose:      game.Purpose{Draws: 1, Scry: 2},
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(Scry{N: 2}, DrawCards{N: 1}),
	})
}
