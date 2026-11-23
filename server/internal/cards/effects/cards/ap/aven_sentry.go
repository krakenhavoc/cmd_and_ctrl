package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Sentry — Creature — Bird Soldier {3}{W}, 3/2 (Dominaria, #5530):
//
//	"Flying"
//
// One printed keyword, which the block check reads (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36a600fe-91b4-46a0-97ee-0cea2b63ec8e",
		Name:            "Aven Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
