package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warren-Scourge Elf — Creature — Elf Warrior {1}{G}, 1/1 (Lorwyn):
//
//	"Protection from Goblins"
//
// Protection from a creature type is a printed keyword the engine honours
// for damage, blocking, targeting and enchanting (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64368768-77b9-4f78-adb7-951ca71dfb28",
		Name:            "Warren-Scourge Elf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Goblins"},
	})
}
