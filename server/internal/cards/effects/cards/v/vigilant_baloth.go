package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigilant Baloth — Creature — Beast {3}{G}{G}, 5/5 (Core Set 2019, #5527):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "caf1bc69-a8d7-454b-9cc6-2e4d003776a8",
		Name:            "Vigilant Baloth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
