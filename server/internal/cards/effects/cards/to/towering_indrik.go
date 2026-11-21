package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Towering Indrik — Creature — Beast {3}{G}, 2/4 (Return to Ravnica, #5566):
//
//	"Reach (This creature can block creatures with flying.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "78bc8b7c-ca94-485c-b7bf-53c3724b7339",
		Name:            "Towering Indrik",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
