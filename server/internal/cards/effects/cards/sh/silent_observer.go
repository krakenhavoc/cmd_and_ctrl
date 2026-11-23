package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silent Observer — Creature — Spirit {3}{U}, 1/5
// (Shadows over Innistrad, #5543):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c377c3e0-2b91-43c6-9f49-5dd6fbf94308",
		Name:            "Silent Observer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
