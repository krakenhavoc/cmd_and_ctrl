package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunspire Griffin — Creature — Griffin {1}{W}{W}, 2/3 (Return to Ravnica, #5566):
//
//	"Flying"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f51ab3d-aae9-4d8d-a881-e8835dcd7866",
		Name:            "Sunspire Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
