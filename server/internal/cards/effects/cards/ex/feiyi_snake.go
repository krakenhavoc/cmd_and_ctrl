package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feiyi Snake — Creature — Snake {1}{G}, 2/1 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Reach"
//
// The keyword is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77b69977-200e-44c2-bfde-228e6501b03a",
		Name:            "Feiyi Snake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
