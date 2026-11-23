package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banehound — Creature — Nightmare Dog {B}, 1/1 (War of the Spark, #5520):
//
//	"Lifelink, haste"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63d846e2-4bef-4911-9508-273c75a4a040",
		Name:            "Banehound",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink", "haste"},
	})
}
