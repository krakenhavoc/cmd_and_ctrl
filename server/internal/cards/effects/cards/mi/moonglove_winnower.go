package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonglove Winnower — Creature — Elf Rogue {3}{B}, 2/3 (Lorwyn):
//
//	"Deathtouch"
//
// Deathtouch is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4870995-f434-4985-a921-1dc9556cee41",
		Name:            "Moonglove Winnower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
