package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Femeref Scouts — Creature — Human Scout {2}{W}, 1/4 (Mirage, S82 OLD batch
// 810, #5633): a vanilla creature, no abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18710b5e-5403-48cd-9541-8154ba0a0b30",
		Name:         "Femeref Scouts",
		Completeness: effects.CompletenessFull,
	})
}
