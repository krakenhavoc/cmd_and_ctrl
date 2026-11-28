package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pegasus Charger — Creature — Pegasus 2/1 {2}{W} (Urza's Saga, S82 OLD b723):
//
//	Flying
//	First strike
//
// Printed keywords only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "81a5ac8d-b904-4755-a5bc-650e95f4138f",
		Name:            "Pegasus Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
