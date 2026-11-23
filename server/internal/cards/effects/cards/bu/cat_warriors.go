package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cat Warriors — Creature — Cat Warrior {1}{G}{G}, 2/2 (Legends, #5639):
//
//	"Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)"
//
// Landwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cee3d8e5-f26f-4f47-8a7b-9ac07ebf9025",
		Name:            "Cat Warriors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
