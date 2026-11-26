package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Duelist — {W} Creature — Human Soldier 1/2: "Protection from
// Vampires"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8e2e1e9-13a6-4cfd-b9ed-2080e984844d",
		Name:            "Midnight Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Vampires"},
	})
}
