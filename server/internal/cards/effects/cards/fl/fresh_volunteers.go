package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fresh Volunteers — Creature — Human Rebel {1}{W}, 2/2 (Mercadian Masques, S82 OLD batch 688, #5619):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd0031c3-ff88-4bc5-82a0-2187c60a7d33",
		Name:         "Fresh Volunteers",
		Completeness: effects.CompletenessFull,
	})
}
