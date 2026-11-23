package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traumatize — Sorcery {3}{U}{U} (EDHREC rank 2745):
//
//	"Target player mills half their library, rounded down."
//
// The self-mill enabler and the mill deck's haymaker in one card. The
// half is read from the library as the spell resolves, rounded down
// (b22MillHalf, Singularity Rupture's body); a target player who left
// the game in response makes the spell do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2ea7d01-6564-4a7f-b935-49a5e3978dac",
		Name:         "Traumatize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.MillHalfOfTargetPlayer,
	})
}
