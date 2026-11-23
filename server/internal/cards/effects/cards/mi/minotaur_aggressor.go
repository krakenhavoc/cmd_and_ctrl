package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minotaur Aggressor — Creature — Minotaur Berserker {6}{R}, 6/2 (Return to Ravnica, #5566):
//
//	"First strike, haste"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7576bfba-eb4f-4d3e-aec5-52daee84d1d6",
		Name:            "Minotaur Aggressor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
