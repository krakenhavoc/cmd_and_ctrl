package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildwood Patrol — Creature — Centaur Scout {2}{G}, 4/2 (Ravnica: Clue Edition):
//
//	"Trample"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4fb9c040-c684-4ba8-970e-58313af5584a",
		Name:            "Wildwood Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
