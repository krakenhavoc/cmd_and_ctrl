package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyline Predator — Creature — Drake {4}{U}{U}, 3/4 (Return to Ravnica, #5566):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.) / Flying"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9bc6d4ac-6fca-4574-9cd8-10a5528575b3",
		Name:            "Skyline Predator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
