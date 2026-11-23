package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peregrine Griffin — Creature — Griffin {4}{W}, 2/4 (Magic 2012, #5573):
//
//	"Flying
//	 First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// Flying and first strike, both printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8848065-7075-41d8-82f6-67ace390625a",
		Name:            "Peregrine Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
