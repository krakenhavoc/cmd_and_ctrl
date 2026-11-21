package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blackcleave Goblin — Creature — Phyrexian Goblin Zombie {3}{B}, 2/1
// (Scars of Mirrodin, #5577):
//
//	"Haste
//	 Infect"
//
// A keyword body: haste and infect (the damage tail, CR 702.90) are the
// engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "156d1a58-c970-4b81-b6c7-5a0caa1699cd",
		Name:            "Blackcleave Goblin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "infect"},
	})
}
