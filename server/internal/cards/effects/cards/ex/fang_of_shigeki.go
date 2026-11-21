package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang of Shigeki — Enchantment Creature — Snake Ninja {G}, 1/1 (Neon Dynasty):
//
//	"Deathtouch"
//
// A keyword-only creature: the printed keyword is the whole card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e187d43-ae75-4589-81fc-476ae95ba5b9",
		Name:            "Fang of Shigeki",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
