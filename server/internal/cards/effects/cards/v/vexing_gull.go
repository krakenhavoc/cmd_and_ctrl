package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vexing Gull — Creature — Bird {2}{U}, 2/2 (Foundations Jumpstart, S82,
// #5929):
//
//	"Flash
//	 Flying"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a36d2a8-6117-4828-bbe2-bc63212d29d8",
		Name:            "Vexing Gull",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
