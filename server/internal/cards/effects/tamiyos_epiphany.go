package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tamiyo's Epiphany — Sorcery {3}{U} (War of the Spark, #5520):
//
//	"Scry 4, then draw two cards."
//
// The draw waits for the scry answer (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48db03c4-c847-4901-8ecb-f480ede94a37",
		Name:         "Tamiyo's Epiphany",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 4, Draws: 2},
		OnResolve:    ResolveSteps(Scry{N: 4}, DrawCards{N: 2}),
	})
}
