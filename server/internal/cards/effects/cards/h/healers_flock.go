package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Healer's Flock — Creature — Bird {W}{W}{W}, 3/3 (Modern Horizons 2,
// #5501):
//
//	"Flying, lifelink"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b3b2fb1-d5fe-4d14-985a-85f13cd60d9c",
		Name:            "Healer's Flock",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
