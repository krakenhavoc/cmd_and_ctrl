package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistral Charger — Creature — Pegasus {1}{W}, 2/1 (Ravnica Remastered):
//
//	"Flying"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a084b37a-bfb2-4dde-9faa-4c55ad894300",
		Name:            "Mistral Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
