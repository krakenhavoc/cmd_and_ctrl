package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Paladin — Creature — Dwarf Knight {1}{W}, 2/2 (Adventures in the Forgotten Realms):
//
//	Lifelink
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e95618c-4f58-4ae9-9e5e-deffd5e210a4",
		Name:            "Steadfast Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
