package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cerulean Wyvern — Creature — Drake {4}{U}, 3/3 (Mirage, S82 OLD b806):
//
//	"Flying, protection from green"
//
// Flying and one protection quality (CR 702.16), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a7f37a3-9e99-4d3d-b705-b074fb3e9b3e",
		Name:            "Cerulean Wyvern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("G")},
	})
}
