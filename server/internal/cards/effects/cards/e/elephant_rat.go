package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elephant-Rat — Creature — Elephant Rat {1}{B}, 1/3 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Menace"
//
// Menace is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84557d9d-ee22-46c9-8f4a-62a440fb302f",
		Name:            "Elephant-Rat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
