package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cerodon Yearling — Creature — Beast {R}{W}, 2/2 (Alara Reborn, #5586):
//
//	"Vigilance, haste"
//
// Two printed keywords the engine enforces.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6db09a0a-6c8e-4829-9b55-00745f0587b5",
		Name:            "Cerodon Yearling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
	})
}
