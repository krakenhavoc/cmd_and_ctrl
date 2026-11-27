package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nath's Buffoon — Creature — Goblin Rogue {1}{B}, 1/1 (Lorwyn):
//
//	"Protection from Elves"
//
// Protection from a creature type is a printed keyword the engine honours
// for damage, blocking, targeting and enchanting (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da404c9c-3fb0-418c-947f-f42d6aff6899",
		Name:            "Nath's Buffoon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Elves"},
	})
}
