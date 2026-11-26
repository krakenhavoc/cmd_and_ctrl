package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Makindi Griffin — Creature — Griffin {3}{W}, 2/4 (Duel Decks: Zendikar vs.
// Eldrazi, #5548):
//
//	"Flying"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af2a951f-7630-48fe-bb37-273b66901eb6",
		Name:            "Makindi Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
