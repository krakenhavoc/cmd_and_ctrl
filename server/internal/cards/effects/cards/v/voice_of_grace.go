package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of Grace — Creature — Angel {3}{W} 2/2 (Urza's Saga, #5624):
//
//	"Flying, protection from black"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7338f964-b7d8-4491-b933-7d8b7d1b56ed",
		Name:            "Voice of Grace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
	})
}
