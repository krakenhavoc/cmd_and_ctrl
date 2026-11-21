package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli's Snubhorn — Creature — Dinosaur {1}{W}, 2/2 (Ixalan, #5532):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8516d8ba-b3f9-4c45-8d70-5dab476e478f",
		Name:            "Huatli's Snubhorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
