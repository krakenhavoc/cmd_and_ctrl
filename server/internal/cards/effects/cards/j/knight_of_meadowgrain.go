package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Meadowgrain — Creature — Kithkin Knight {W}{W}, 2/2 (Lorwyn):
//
//	"First strike
//	 Lifelink"
//
// Both keywords are the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4e6c71e8-aeaf-404e-84c2-c79f2a1cbe4b",
		Name:            "Knight of Meadowgrain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
	})
}
