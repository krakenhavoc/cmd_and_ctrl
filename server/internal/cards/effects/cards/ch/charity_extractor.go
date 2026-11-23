package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charity Extractor — Creature — Human Knight {3}{B}, 1/5 (War of the Spark, #5520):
//
//	"Lifelink"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f18a1a99-8363-4edf-9807-2e23fa007ce8",
		Name:            "Charity Extractor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
