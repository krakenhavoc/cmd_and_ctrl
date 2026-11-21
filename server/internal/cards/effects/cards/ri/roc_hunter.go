package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roc Hunter — Creature — Human Soldier {1}{R}, 3/1 (The Brothers' War, #5482):
//
//	"Reach"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ed811b7-6ead-4906-9e40-9096b735e271",
		Name:            "Roc Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
