package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colorful Feiyi Sparrow — Creature — Bird {1}{W}, 1/3 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Flying"
//
// The keyword is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82657b8b-de2d-46ab-ab56-e2790e326d95",
		Name:            "Colorful Feiyi Sparrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
