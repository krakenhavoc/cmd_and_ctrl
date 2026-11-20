package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turtle-Seals — Creature — Turtle Seal {3}{U}, 2/4 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Vigilance"
//
// Vigilance is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0172cad-8b93-4d1a-b442-5eced3212b17",
		Name:            "Turtle-Seals",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
