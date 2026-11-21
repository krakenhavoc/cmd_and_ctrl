package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Cancrix — Creature — Crab {4}{U}, 2/5 (Magic 2011, #5579):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "445bb343-9758-459b-970f-47f1705f1e55",
		Name:         "Armored Cancrix",
		Completeness: effects.CompletenessFull,
	})
}
