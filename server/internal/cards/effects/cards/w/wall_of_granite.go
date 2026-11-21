package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Granite — Creature — Wall {2}{R}, 0/7 (Portal, #5630):
//
//	"Defender"
//
// Defender is an engine keyword carried as printed data: it can't
// attack (CR 702.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8445094f-008b-491a-977c-e8582d5ab72c",
		Name:            "Wall of Granite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
