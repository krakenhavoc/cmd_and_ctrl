package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tundra Wall — Creature — Wall {1}{W}, 0/4 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Defender"
//
// Defender is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68df4c46-513a-4d75-9d4b-ac95d45be255",
		Name:            "Tundra Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
