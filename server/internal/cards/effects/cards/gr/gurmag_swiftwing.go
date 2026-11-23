package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gurmag Swiftwing — Creature — Bat {1}{B}, 1/2 (Ravnica: Clue Edition,
// #5929):
//
//	"Flying, first strike, haste"
//
// Three printed keywords and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc1256e9-eb92-43b6-abfc-ce19d910a929",
		Name:            "Gurmag Swiftwing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "haste"},
	})
}
