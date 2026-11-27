package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oakgnarl Warrior — Creature — Treefolk Warrior {5}{G}{G}, 5/7 (Lorwyn):
//
//	"Vigilance, trample"
//
// Both keywords are the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cafdf77-5500-4471-a848-89863eea2a35",
		Name:            "Oakgnarl Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
	})
}
