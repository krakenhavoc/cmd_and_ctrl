package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Phalanx — Creature — Human Soldier {4}{W}, 2/4 (Fallen Empires, S82
// OLD b854, #5637):
//
//	"Banding"
//
// Banding rides PrintedKeywords; the engine owns the rest (CR 702.22).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1160b8d-c4e4-4869-b7a5-c6a826eb1ba4",
		Name:            "Icatian Phalanx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
