package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightguard Patrol — Creature — Human Soldier {2}{W}, 2/1 (Jumpstart 2022):
//
//	"First strike, vigilance"
//
// Two printed keywords, nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb482f90-5624-4e2c-9916-85fb09a3639a",
		Name:            "Nightguard Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
	})
}
