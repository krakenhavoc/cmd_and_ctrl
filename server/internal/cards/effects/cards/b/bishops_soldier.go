package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bishop's Soldier — Creature — Vampire Soldier {1}{W}, 2/2 (Foundations, #4465):
//
//	"Lifelink"
//
// Lifelink is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84487586-4f71-457a-b983-4f1f7a71deef",
		Name:            "Bishop's Soldier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
