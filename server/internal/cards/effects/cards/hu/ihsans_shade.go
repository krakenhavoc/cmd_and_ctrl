package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ihsan's Shade — Legendary Creature — Shade Knight {3}{B}{B}{B}, 5/5
// (Homelands, S82 OLD b826, #5635):
//
//	"Protection from white"
//
// Protection from white is a printed keyword (CR 702.16b): it can't be
// damaged, enchanted, blocked or targeted by anything white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5eb29367-4ec3-4bd2-887e-13eeeecae8fa",
		Name:            "Ihsan's Shade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
	})
}
