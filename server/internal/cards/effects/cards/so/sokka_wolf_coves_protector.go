package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokka, Wolf Cove's Protector — Legendary Creature — Human Warrior Ally {2}{W}, 3/3 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Vigilance"
//
// Vigilance is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dd518f9-1ddd-4950-add3-fc21f675f4dc",
		Name:            "Sokka, Wolf Cove's Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
