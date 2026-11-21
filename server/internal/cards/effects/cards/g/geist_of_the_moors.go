package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist of the Moors — Creature — Spirit {1}{W}{W}, 3/1 (Magic 2015,
// #5556):
//
//	"Flying"
//
// Armored Pegasus's printed flying, carried as data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c9d814b-4257-47bc-a1df-c3480bc038ba",
		Name:            "Geist of the Moors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
