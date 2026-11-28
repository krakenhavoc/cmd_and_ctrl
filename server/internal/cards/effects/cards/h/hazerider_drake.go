package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hazerider Drake — Creature — Drake {2}{W}{U}, 2/3 (Mirage, S82 OLD b807):
//
//	"Flying, protection from red"
//
// Flying and one protection quality (CR 702.16), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9be76a3-e14f-48dd-8824-17d5708ac7e0",
		Name:            "Hazerider Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("R")},
	})
}
