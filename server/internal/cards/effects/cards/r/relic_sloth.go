package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Sloth — Creature — Sloth Beast {3}{R}{W}, 4/4 (Strixhaven: School
// of Mages, #5502):
//
//	"Vigilance
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)"
//
// Two printed keywords the combat engine reads (Masked Bandits' pair).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb9b0655-3ca3-4d12-8665-54ef0c2c4de2",
		Name:            "Relic Sloth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
	})
}
