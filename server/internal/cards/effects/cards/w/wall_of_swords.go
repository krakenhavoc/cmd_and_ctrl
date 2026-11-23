package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Swords — Creature — Wall {3}{W}, 3/5 (Portal, #5630):
//
//	"Defender (This creature can't attack.)
//	Flying"
//
// Defender and flying are engine keywords carried as printed data: it
// can't attack (CR 702.3b) and only a creature with flying or reach can
// block it (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb098958-50d3-4476-ba74-382033703ff9",
		Name:            "Wall of Swords",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
	})
}
