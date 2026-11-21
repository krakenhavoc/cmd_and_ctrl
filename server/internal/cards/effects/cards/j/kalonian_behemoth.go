package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalonian Behemoth — Creature — Beast {5}{G}{G}, 9/9 (Magic 2010, #5585):
//
//	"Shroud (This creature can't be the target of spells or abilities.)"
//
// Deadly Insect's shroud (CR 702.18a), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68d0b5c0-df6a-4eea-8dea-20358270f916",
		Name:            "Kalonian Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
