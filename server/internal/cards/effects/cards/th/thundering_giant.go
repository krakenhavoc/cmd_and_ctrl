package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Giant — Creature — Giant {3}{R}{R}, 4/3 (Magic 2015, #5556):
//
//	"Haste"
//
// A vanilla haste creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d23329e2-4134-4e49-b9b6-5febef2bce31",
		Name:            "Thundering Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
