package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Skyhunter — Creature — Cat Knight {W}{W}, 2/2 (Foundations, #4465):
//
//	"Flying"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "517c7295-8ba2-47a6-a1c6-aee722f8ca88",
		Name:            "Leonin Skyhunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
