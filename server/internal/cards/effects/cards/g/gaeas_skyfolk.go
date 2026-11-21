package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Skyfolk — Creature — Elf Merfolk {G}{U}, 2/2 (Apocalypse,
// #5614):
//
//	"Flying"
//
// One printed keyword on the PrintedKeywords slot, so the engine's own
// flying rule (CR 702.9b) decides what can block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd0da475-55d0-462e-bfac-6eeb5d39a099",
		Name:            "Gaea's Skyfolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
