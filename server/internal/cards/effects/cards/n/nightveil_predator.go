package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightveil Predator — Creature — Vampire {U}{U}{B}{B}, 3/3 (Guilds of Ravnica, #5525):
//
//	"Flying, deathtouch
//	 Hexproof"
//
// Three engine keywords carried as printed data: flying, deathtouch and
// hexproof (CR 702.11b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b94591b8-2e99-4983-8f36-9b52207bfd2e",
		Name:            "Nightveil Predator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "hexproof"},
	})
}
