package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rib Cage Spider — Creature — Spider {2}{G}, 1/4 (Prophecy, S82 OLD batch 673,
// #5617):
//
//	"Reach (This creature can block creatures with flying.)"
//
// A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "906cba93-3dac-4720-a482-987cf1b4e786",
		Name:            "Rib Cage Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
