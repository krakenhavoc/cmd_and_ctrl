package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shepherd of the Lost — Creature — Angel {4}{W}, 3/3 (Zendikar, #5583):
//
//	"Flying, first strike, vigilance"
//
// Three printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8c70f2f-985c-4982-97eb-99da3810e048",
		Name:            "Shepherd of the Lost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance"},
	})
}
