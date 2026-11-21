package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden-Tail Disciple — Enchantment Creature — Fox Monk {2}{W}, 2/3 (Neon Dynasty):
//
//	"Lifelink"
//
// A keyword-only creature: the printed keyword is the whole card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "316f7091-0c75-4510-a89a-6b58bb7f0627",
		Name:            "Golden-Tail Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
