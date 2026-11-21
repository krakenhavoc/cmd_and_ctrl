package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underdark Basilisk — Creature — Basilisk {1}{G}, 1/2 (Adventures in the Forgotten Realms):
//
//	Deathtouch
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b61b6da-44ed-4766-8134-fca37afd3f2a",
		Name:            "Underdark Basilisk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
