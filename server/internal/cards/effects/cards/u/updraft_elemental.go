package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Updraft Elemental — Creature — Elemental {2}{U}, 1/4 (Dragons of Tarkir,
// #5551):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a33982e4-d7d2-4319-bf04-534b8f253dfb",
		Name:            "Updraft Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
