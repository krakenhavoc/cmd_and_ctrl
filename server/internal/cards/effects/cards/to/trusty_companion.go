package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trusty Companion — Creature — Hyena {1}{W}, 3/2 (Kaladesh, S82):
//
//	"Vigilance
//	 This creature can't attack alone."
//
// Pet Sehlat's attack limit (CantAttackAlone, CR 506.5, 508.1c) with
// vigilance read off the printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd8c633c-ba99-47ac-a1d1-9e15aa948933",
		Name:            "Trusty Companion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AttackLimits:    []game.AttackLimit{effects.CantAttackAlone()},
	})
}
