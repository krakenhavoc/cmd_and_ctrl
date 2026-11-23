package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Concordia Pegasus — Creature — Pegasus {1}{W}, 1/3 (Return to Ravnica, #5566):
//
//	"Flying"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93c35f9e-3346-4269-b1af-df9f216bfd7a",
		Name:            "Concordia Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
