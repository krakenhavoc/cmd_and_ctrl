package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Escort — Creature — Human Soldier {2}{W}{W} 2/3 (Alliances, S82
// OLD batch 820): "Banding".
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c6c1604-0724-40b1-9df6-71e57e5063c9",
		Name:            "Kjeldoran Escort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
