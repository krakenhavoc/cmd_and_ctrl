package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootbreaker Wurm — Creature — Wurm {5}{G}{G}, 6/6 (Tempest, S82 OLD
// batch 758, #5628):
//
//	"Trample"
//
// A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3edbb47-6892-4853-badc-cc01499d4e55",
		Name:            "Rootbreaker Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
