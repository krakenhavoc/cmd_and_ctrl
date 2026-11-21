package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hostile Minotaur — Creature — Minotaur {3}{R}, 3/3 (Core Set 2019, #5527):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your control.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e050a0a3-f48e-4cba-9cd4-692a92261c9f",
		Name:            "Hostile Minotaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
