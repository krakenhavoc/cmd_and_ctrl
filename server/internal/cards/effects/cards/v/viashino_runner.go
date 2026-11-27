package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Runner — Creature — Lizard {3}{R} 3/2 (Urza's Saga, #5624):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac7e317b-387c-403f-bcf1-403f96302f21",
		Name:            "Viashino Runner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
