package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rushwood Dryad — Creature — Dryad {1}{G}, 2/1 (Eighth Edition, #5607):
//
//	"Forestwalk"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb1963f9-261c-4d3f-a401-90aaced77e7f",
		Name:            "Rushwood Dryad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
