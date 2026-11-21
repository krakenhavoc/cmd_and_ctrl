package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brazen Scourge — Creature — Gremlin {1}{R}{R}, 3/3 (Foundations, #4465):
//
//	"Haste"
//
// Haste is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "422ef4ec-863a-450e-9577-a0707bc16532",
		Name:            "Brazen Scourge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
