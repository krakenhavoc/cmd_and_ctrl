package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivisurgeon's Insight — Sorcery {3}{U}{U} (Phyrexia: All Will Be One,
// #5479):
//
//	"Draw three cards. Proliferate."
//
// Two steps, so the proliferate waits for a prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf6ff089-63d7-4a98-b95d-2837c017a866",
		Name:         "Vivisurgeon's Insight",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve:    ResolveSteps(DrawCards{N: 3}, Proliferate{}),
	})
}
