package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nacatl Savage — Creature — Cat Warrior {1}{G}, 2/1 (Conflux,
// #5587):
//
//	"Protection from artifacts"
//
// Protection from artifacts is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "475412c9-8960-4085-bd8f-5f8a7c7485b9",
		Name:            "Nacatl Savage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from artifacts"},
	})
}
