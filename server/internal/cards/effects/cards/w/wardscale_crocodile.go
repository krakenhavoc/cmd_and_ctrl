package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wardscale Crocodile — Creature — Crocodile {4}{G}, 5/3 (War of the Spark, #5520):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86f725f6-b48b-4dcc-9a02-e258a226e31c",
		Name:            "Wardscale Crocodile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
