package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sungrace Pegasus — {1}{W} Creature — Pegasus 1/2 (Foundations Jumpstart):
//
//	"Flying
//	 Lifelink"
//
// Both keywords ride PrintedKeywords. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67f59250-d46a-4637-9493-165d23278817",
		Name:            "Sungrace Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
