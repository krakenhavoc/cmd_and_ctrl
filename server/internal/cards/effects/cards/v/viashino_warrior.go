package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Warrior — Creature — Lizard Warrior {3}{R}, 4/2 (Mirage, S82 OLD
// batch 810, #5633): a vanilla creature, no abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f46e06e2-1658-4bec-b94e-91935cc667fb",
		Name:         "Viashino Warrior",
		Completeness: effects.CompletenessFull,
	})
}
