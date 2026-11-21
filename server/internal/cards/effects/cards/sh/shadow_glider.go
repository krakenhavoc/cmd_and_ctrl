package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Glider — Creature — Kor Soldier {2}{W}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b92e0156-ea0f-4b13-b3bc-7e84a31445fc",
		Name:            "Shadow Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
