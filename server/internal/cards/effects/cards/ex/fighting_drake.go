package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fighting Drake — Creature — Drake {2}{U}{U}, 2/4 (Eighth Edition, #5607):
//
//	"Flying"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b13cc6be-ee35-491e-a437-f7fc191f24c7",
		Name:            "Fighting Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
