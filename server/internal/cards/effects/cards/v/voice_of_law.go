package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of Law — Creature — Angel {3}{W} 2/2 (Urza's Saga, #5624):
//
//	"Flying, protection from red"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "535007c5-03a8-495f-a132-37960f179f96",
		Name:            "Voice of Law",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
