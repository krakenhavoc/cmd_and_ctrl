package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mist Leopard — Creature — Cat {3}{G}, 3/2 (Magic 2010, #5585):
//
//	"Shroud (This creature can't be the target of spells or abilities.)"
//
// Deadly Insect's shroud (CR 702.18a), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49cade2f-92ab-48c8-aed3-d52ab3263506",
		Name:            "Mist Leopard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
