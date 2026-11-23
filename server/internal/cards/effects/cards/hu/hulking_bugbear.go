package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hulking Bugbear — Creature — Goblin {1}{R}{R}, 3/3 (Adventures in the Forgotten Realms):
//
//	Haste
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c336eb5b-e185-4d4b-b106-3a641ba9aad1",
		Name:            "Hulking Bugbear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
