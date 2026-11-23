package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predator Dragon — "Flying, haste. Devour 2"
//
// Devour 2 is devour.go's: sacrifice any number of creatures as it
// enters, two +1/+1 counters for each. No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef048ecc-581d-4a17-91fd-edf7d7733ef9",
		Name:            "Predator Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Replacements:    []game.ReplacementEffect{effects.Devour("Predator Dragon", 2)},
	})
}
