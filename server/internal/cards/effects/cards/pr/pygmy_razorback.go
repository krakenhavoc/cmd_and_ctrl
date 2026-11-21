package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pygmy Razorback — Creature — Boar {1}{G}, 2/1 (Prophecy, #5617):
//
//	"Trample"
//
// One printed keyword, which the combat damage assignment reads
// (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2bdd82e7-cbbf-4f12-bcb7-b276df021351",
		Name:            "Pygmy Razorback",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
