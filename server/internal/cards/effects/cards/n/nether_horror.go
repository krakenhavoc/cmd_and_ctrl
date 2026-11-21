package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nether Horror — Creature — Horror {3}{B}, 4/2 (Magic 2011, #5579):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4647f16-b287-46a2-9c38-93881c290f76",
		Name:         "Nether Horror",
		Completeness: effects.CompletenessFull,
	})
}
