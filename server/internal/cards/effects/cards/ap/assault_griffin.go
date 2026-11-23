package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assault Griffin — Creature — Griffin {3}{W}, 3/2 (Magic 2011, #5579):
//
//	"Flying"
//
// Flying is the printed keyword the block check reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "802b9333-c256-410a-a635-a22bfeda679f",
		Name:            "Assault Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
