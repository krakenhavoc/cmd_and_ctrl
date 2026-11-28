package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Air — Creature — Wall {1}{U}{U}, 1/5 (Eighth Edition, #5607):
//
//	"Defender, flying"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2d3da40-e2f7-4480-9da9-33019d6f4071",
		Name:            "Wall of Air",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
	})
}
