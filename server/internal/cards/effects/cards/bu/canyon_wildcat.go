package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canyon Wildcat — Creature — Cat {1}{R}, 2/1 (Eighth Edition, #5607):
//
//	"Mountainwalk"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbf7b3ef-1ca4-4b0d-bec2-f5e60b150013",
		Name:            "Canyon Wildcat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
