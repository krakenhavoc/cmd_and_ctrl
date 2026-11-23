package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bull Hippo — Creature — Hippo {3}{G}, 3/3 (Portal, #5630):
//
//	"Islandwalk"
//
// Islandwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls an Island (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d62a448-b6a5-43b1-a281-9e9361a5524a",
		Name:            "Bull Hippo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
