package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitterbow Sharpshooters — Creature — Jackal Archer {4}{G}, 4/4
// (Hour of Devastation, #5534):
//
//	"Reach, vigilance"
//
// Every word is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "76f22670-6806-4124-8bbb-822ed33d2bda",
		Name:            "Bitterbow Sharpshooters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
	})
}
