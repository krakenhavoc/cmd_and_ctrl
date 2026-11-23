package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Two-Headed Zombie — Creature — Zombie {3}{B}, 4/2 (Core Set 2019, #5527):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e280770-9a8f-4445-84d2-3b9cac659136",
		Name:            "Two-Headed Zombie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
