package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Spirit — Creature — Spirit {1}{W}{U}, 2/2 (Conspiracy, S82 OLD batch
// 364, #5557):
//
//	"Flying, first strike"
//
// Two printed keywords, read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc7e288f-ade2-4997-b3fe-f873c26990b0",
		Name:            "Sky Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
