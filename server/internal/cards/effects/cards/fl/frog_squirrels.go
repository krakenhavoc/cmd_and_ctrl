package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frog-Squirrels — Creature — Frog Squirrel {1}{G}, 2/2 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Reach (This creature can block creatures with flying.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b8563788-debc-4f97-adf4-b186bf3960e2",
		Name:            "Frog-Squirrels",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
