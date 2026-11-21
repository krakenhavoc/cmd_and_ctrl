package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Raptor — Creature — Bird {W}, 1/2 (Kaldheim, #5505):
//
//	"Flying, first strike"
//
// A keyword body: the keywords ride PrintedKeywords and the engine's
// combat and damage rules do the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68a99e50-0e98-480b-ad25-aa2c226d24ca",
		Name:            "Battlefield Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
