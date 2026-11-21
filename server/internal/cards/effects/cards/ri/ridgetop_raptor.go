package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ridgetop Raptor — Creature — Dinosaur Beast {3}{R}, 2/1 (Dominaria
// Remastered, #5948):
//
//	"Double strike"
//
// The keyword rides PrintedKeywords (CR 702.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "600a8c9e-e158-4fff-8c4d-c3a1bfd006e8",
		Name:            "Ridgetop Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
