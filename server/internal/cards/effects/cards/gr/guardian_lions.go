package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Lions — Creature — Cat {4}{W}, 1/6 (Magic 2013, #5568):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// A keyword body: vigilance (CR 702.20).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9291f31d-6953-4fe1-9a88-b8dd09fde89c",
		Name:            "Guardian Lions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
