package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Zombie — Creature — Merfolk Zombie {U}{B}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from green"
//
// Protection from green is a printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58e4f299-bb32-4727-8498-a364166adc38",
		Name:            "Vodalian Zombie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from green"},
	})
}
