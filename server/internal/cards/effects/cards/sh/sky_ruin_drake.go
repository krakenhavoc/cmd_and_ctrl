package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Ruin Drake — Creature — Drake {4}{U}, 2/5 (Zendikar, #5583):
//
//	"Flying"
//
// Flying is the printed keyword (CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9b7881b-915a-42f1-bd2f-0aaa900a8981",
		Name:            "Sky Ruin Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
