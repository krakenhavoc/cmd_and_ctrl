package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plover Knights — Creature — Kithkin Knight {3}{W}{W}, 3/3 (Lorwyn):
//
//	"Flying, first strike"
//
// Both keywords are the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4a99095-38cc-4fd3-b3da-c72a6ccceb25",
		Name:            "Plover Knights",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
