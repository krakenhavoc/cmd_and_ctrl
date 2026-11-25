package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormcloud Spirit — Creature — Spirit {3}{U}{U}, 4/4 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Flying"
//
// The keyword is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0dc8bdbb-996b-47d7-a5e5-7bdb6d9280fe",
		Name:            "Stormcloud Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
