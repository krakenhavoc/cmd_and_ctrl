package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steeple Roc — Creature — Bird {4}{W}, 3/1:
//
//	"Flying, first strike"
//
// Flying and first strike, both printed keywords the engine reads in
// combat (CR 702.9b, CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1ba6053-14e0-46b0-b760-b72e4fa0c5f5",
		Name:            "Steeple Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
