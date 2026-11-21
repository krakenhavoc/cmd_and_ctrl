package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rorix Bladewing — Legendary Creature — Dragon {3}{R}{R}{R}, 6/5
// (Commander Masters):
//
//	"Flying, haste"
//
// Two printed keywords, nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48fe0f52-1676-49d5-acf6-5476081e8e1b",
		Name:            "Rorix Bladewing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
