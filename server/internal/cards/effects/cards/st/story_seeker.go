package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Story Seeker — Creature — Dwarf Cleric {1}{W}, 2/2 (Kaldheim, #5505):
//
//	"Lifelink"
//
// A keyword body: the keywords ride PrintedKeywords and the engine's
// combat and damage rules do the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd7abcee-d4d3-4cf6-8385-c6ad7ac218e4",
		Name:            "Story Seeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
