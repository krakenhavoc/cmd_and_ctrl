package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breakneck Berserker — Creature — Dwarf Berserker {2}{R}, 3/2 (Kaldheim, #5505):
//
//	"Haste"
//
// A keyword body: the keywords ride PrintedKeywords and the engine's
// combat and damage rules do the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9f18a580-1c30-4f97-be06-098a763dfda8",
		Name:            "Breakneck Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
