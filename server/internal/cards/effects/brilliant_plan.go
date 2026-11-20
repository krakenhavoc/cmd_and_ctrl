package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brilliant Plan — Sorcery {4}{U} (Portal Three Kingdoms, #5622):
//
//	"Draw three cards."
//
// Harmonize's body (DrawCards for its controller): one draw of three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d83e1a42-11d1-412a-b66c-850c7a528777",
		Name:         "Brilliant Plan",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve:    ResolveSteps(DrawCards{N: 3}),
	})
}
