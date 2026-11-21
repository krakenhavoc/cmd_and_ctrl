package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguinary Mage — Creature — Vampire Wizard {1}{R}, 1/3
// (Shadows over Innistrad, #5543):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Prowess is an engine keyword carried as printed data (CR 702.108), as on
// Bloodfire Expert.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3fb0388-bfa8-4d22-831c-40a81a11d8c6",
		Name:            "Sanguinary Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
