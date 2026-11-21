package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilacorn — Creature — Lizard {B}, 1/1 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1551cd95-1e68-404f-be1d-cc9c7f447210",
		Name:            "Gilacorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
