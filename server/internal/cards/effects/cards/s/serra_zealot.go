package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra Zealot — Creature — Human Soldier {W} 1/1 (Urza's Saga, #5624):
//
//	"First strike"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "989a7353-8d3d-4ea2-ab5e-8535d95dddae",
		Name:            "Serra Zealot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
