package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hawkeater Moth — Creature — Insect {3}{G} 1/2 (Urza's Saga, #5624):
//
//	"Flying
//	 Shroud (This creature can't be the target of spells or abilities.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3f4bc4c-e747-4bc1-bfba-4ecfa0a5d17f",
		Name:            "Hawkeater Moth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "shroud"},
	})
}
