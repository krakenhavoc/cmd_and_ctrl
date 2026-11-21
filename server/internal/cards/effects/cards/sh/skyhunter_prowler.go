package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyhunter Prowler — Creature — Cat Knight {2}{W}, 1/3 (Jumpstart 2022):
//
//	"Flying, vigilance"
//
// Two printed keywords, nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e0d13b0-d5d0-4c5f-945e-5cda7d980670",
		Name:            "Skyhunter Prowler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
