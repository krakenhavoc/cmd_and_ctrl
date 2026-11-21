package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bay Falcon — Creature — Bird {1}{U}, 1/1 (Mirage, #5633):
//
//	"Flying, vigilance"
//
// Two printed keywords: flying for blocks (CR 702.9b), vigilance for
// attacks (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "709e99f1-288b-49e0-80a1-f817b457a2cf",
		Name:            "Bay Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
