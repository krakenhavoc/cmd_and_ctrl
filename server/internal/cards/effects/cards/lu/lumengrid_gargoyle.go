package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumengrid Gargoyle — Artifact Creature — Gargoyle {6}, 4/4 (Mirrodin Besieged,
// #5576):
//
//	"Flying"
//
// Flying is a printed keyword read at the block declaration (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1070ede7-c81d-4c77-a1a9-6f243432a640",
		Name:            "Lumengrid Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
