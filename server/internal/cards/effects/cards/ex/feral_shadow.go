package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Shadow — Creature — Nightstalker {2}{B}, 2/1 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf26dcb4-e181-4ba5-bc03-b56f57032b85",
		Name:            "Feral Shadow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
