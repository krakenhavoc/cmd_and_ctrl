package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunblade Angel — Creature — Angel {5}{W}, 3/3 (War of the Spark, #5520):
//
//	"Flying, first strike, vigilance, lifelink"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be37ef9d-955a-417e-bde8-bb9e0bdf3612",
		Name:            "Sunblade Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance", "lifelink"},
	})
}
