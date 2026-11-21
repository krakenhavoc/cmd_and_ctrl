package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wilt-Leaf Cavaliers — Creature — Elf Knight {G/W}{G/W}{G/W}, 3/4
// (Shadowmoor, #5591):
//
//	"Vigilance"
//
// Vigilance is a printed keyword (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69ae765d-03c9-46d9-bdad-8e5936df9616",
		Name:            "Wilt-Leaf Cavaliers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
