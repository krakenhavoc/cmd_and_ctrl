package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Billyrider — Creature — Kithkin Knight 1/3 {2}{W} (March of the Machine, S82):
//
//	"Double strike"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "142d2063-d692-42a6-b828-42d998914589",
		Name:            "Kithkin Billyrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
