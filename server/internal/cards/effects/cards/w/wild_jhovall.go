package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Jhovall — Creature — Cat {3}{R}, 3/3 (Mercadian Masques, S82 OLD batch 688, #5619):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8db78c62-e0b9-4ed7-b29f-238f23f79bc4",
		Name:         "Wild Jhovall",
		Completeness: effects.CompletenessFull,
	})
}
