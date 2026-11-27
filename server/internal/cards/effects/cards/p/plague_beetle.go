package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Beetle — Creature — Insect {B}, 1/1 (Eighth Edition, #5607):
//
//	"Swampwalk"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96415c51-2d44-4a96-9caa-9c4fbde59fca",
		Name:            "Plague Beetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
