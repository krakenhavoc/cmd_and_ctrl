package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drakewing Krasis — Creature — Lizard Drake {1}{G}{U}, 3/1 (Gatecrash, #5565):
//
//	"Flying, trample"
//
// Flying and trample are engine keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "782e8499-0714-4cee-869e-8ebd7dad0242",
		Name:            "Drakewing Krasis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
}
