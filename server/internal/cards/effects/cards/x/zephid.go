package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephid — Creature — Illusion {4}{U}{U} 3/4 (Urza's Saga, #5624):
//
//	"Flying
//	 Shroud (This creature can't be the target of spells or abilities.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "88aa710c-26ed-490d-9a4b-4a2b48df1733",
		Name:            "Zephid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "shroud"},
	})
}
