package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Ragemutt — Creature — Elemental Dog {3}{B}{R}, 3/3 (Return to Ravnica, #5566):
//
//	"Lifelink, haste"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f4c7419-bbbe-4ddc-83ae-fd9ea3e86c3d",
		Name:            "Rakdos Ragemutt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink", "haste"},
	})
}
