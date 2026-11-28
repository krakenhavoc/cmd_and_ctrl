package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benthic Behemoth — Creature — Serpent {5}{U}{U}{U}, 7/6 (Tempest, S82 OLD
// batch 767, #5628):
//
//	"Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)"
//
// A printed landwalk keyword (CR 702.14): the block legality check looks for an
// Island among the defending player's lands as the blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aea98950-536e-451d-bd25-ae6e3e459ffb",
		Name:            "Benthic Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
