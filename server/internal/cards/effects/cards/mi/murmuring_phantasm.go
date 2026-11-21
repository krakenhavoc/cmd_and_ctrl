package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murmuring Phantasm — Creature — Spirit {1}{U}, 0/5:
//
//	"Defender"
//
// Defender, read at the attack declaration (CR 702.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35e684fd-69e3-4137-bf15-94f705191f5e",
		Name:            "Murmuring Phantasm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
