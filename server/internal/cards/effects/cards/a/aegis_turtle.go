package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis Turtle — Creature — Turtle {U}, 0/5 (Foundations, #4465):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f84c30e7-2ea8-43ff-9356-1f907558cfd9",
		Name:         "Aegis Turtle",
		Completeness: effects.CompletenessFull,
	})
}
