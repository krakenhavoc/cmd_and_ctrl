package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mammoth Spider — Creature — Spider {4}{G}, 3/5 (Dominaria, #5530):
//
//	"Reach (This creature can block creatures with flying.)"
//
// One printed keyword: it can block a creature with flying (CR 702.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df77cb98-e3c6-4927-99d5-c3380641f356",
		Name:            "Mammoth Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
