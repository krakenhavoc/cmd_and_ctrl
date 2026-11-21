package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squirrelanoids — {B} Creature — Squirrel Mutant 1/1 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Deathtouch"
//
// Typhoid Rats's printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "315ddc7a-c39c-43d9-b134-23acbd865e3e",
		Name:            "Squirrelanoids",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
