package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyhunter Skirmisher — Creature — Cat Knight {1}{W}{W}, 1/1 (Fifth Dawn, #5604):
//
//	"Flying, double strike"
//
// Two printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c07d09e-e127-4573-b827-6c50246f7a31",
		Name:            "Skyhunter Skirmisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
	})
}
