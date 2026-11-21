package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gold-Forged Sentinel — Artifact Creature — Chimera {6}, 4/4 (Journey into Nyx, #5558):
//
//	"Flying"
//
// Ornithopter's printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c338fb29-2dd2-4f76-ba3d-fa6770adcce8",
		Name:            "Gold-Forged Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
