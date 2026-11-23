package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hitchclaw Recluse — Creature — Spider {2}{G}, 1/4 (Magic Origins, #5549):
//
//	"Reach"
//
// Giant Spider's shape: reach feeds Game.CanBlockLocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "241ec1fc-cc50-468e-8ee3-f95a82a3e07b",
		Name:            "Hitchclaw Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
