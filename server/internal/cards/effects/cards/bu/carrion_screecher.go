package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrion Screecher — Creature — Zombie Bird {3}{B}, 3/1
// (Hour of Devastation, #5534):
//
//	"Flying"
//
// Every word is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b813e46-34f6-4918-97a7-a4b2a2464e86",
		Name:            "Carrion Screecher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
