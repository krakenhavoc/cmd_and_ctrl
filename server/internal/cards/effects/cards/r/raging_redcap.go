package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Redcap — Creature — Goblin Knight {2}{R}, 1/2 (Foundations, #4465):
//
//	"Double strike"
//
// Double strike is the canonical keyword token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "679841ad-f3ac-48e2-ba11-2cf6d78642b4",
		Name:            "Raging Redcap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
