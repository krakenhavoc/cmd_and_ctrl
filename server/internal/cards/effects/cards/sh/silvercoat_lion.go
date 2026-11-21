package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silvercoat Lion — Creature — Cat {1}{W}, 2/2 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3c0b11e-7340-4e0a-98b7-91b38439d4f9",
		Name:         "Silvercoat Lion",
		Completeness: effects.CompletenessFull,
	})
}
