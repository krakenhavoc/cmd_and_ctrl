package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needlepeak Spider — Creature — Spider {3}{R}, 4/2 (Planar Chaos, #5595):
//
//	"Reach"
//
// Reach is an engine keyword carried as printed data (CR 702.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "169429d8-aa98-46a8-8681-521c65dafbb4",
		Name:            "Needlepeak Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
