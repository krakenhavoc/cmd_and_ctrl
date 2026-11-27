package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commander Eesha — Legendary Creature — Bird Soldier {2}{W}{W}, 2/4 (Judgment):
//
//	"Flying, protection from creatures"
//
// Beloved Chaplain's protection from a card type (CR 702.16a) with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9b9edb2-62f5-499c-b2fc-359dc673c4b4",
		Name:            "Commander Eesha",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from creatures"},
	})
}
