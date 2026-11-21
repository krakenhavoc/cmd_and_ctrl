package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Assassin (Streets of New Capenna, #5491):
//
//	Flying, deathtouch
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f394d2bf-89aa-44c1-b473-8c7a8874a9a3",
		Name:            "Midnight Assassin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
