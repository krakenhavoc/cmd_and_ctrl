package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloomwidow — Creature — Spider {2}{G}, 3/3 (Shadows over Innistrad, S82 OLD
// batch 290, #5543):
//
//	"Reach
//	 This creature can block only creatures with flying."
//
// Reach is a printed keyword; the block rule is the shared
// NemBlocksOnlyFlyers (Belbe's Percher), read at the declaration of blockers
// (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5139376-cdb1-4f1e-b417-b956d6b713b9",
		Name:            "Gloomwidow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
