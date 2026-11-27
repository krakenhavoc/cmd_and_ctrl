package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weatherseed Faeries — Creature — Faerie {2}{U} 2/1:
// "Flying, protection from red".
func init() {
	effects.Register(effects.Spec{
		OracleID:        "904b2caf-e083-4917-8841-95f48d90dc0c",
		Name:            "Weatherseed Faeries",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
