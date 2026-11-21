package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garruk's Companion — Creature — Beast {G}{G}, 3/2 (Magic 2011, #5579):
//
//	"Trample"
//
// Trample is the printed keyword the combat damage assignment reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32fc8fa7-7e0a-4d4e-85cf-2f98b3fc6ecf",
		Name:            "Garruk's Companion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
