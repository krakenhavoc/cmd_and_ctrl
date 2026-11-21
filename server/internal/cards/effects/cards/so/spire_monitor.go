package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Monitor — Creature — Phyrexian Drake {4}{U}, 3/3 (New Phyrexia, #5575):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Flying"
//
// Wind Strider's two keywords (CR 702.8, CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "447b9412-2649-4548-8e9b-eed5e92ad92a",
		Name:            "Spire Monitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
