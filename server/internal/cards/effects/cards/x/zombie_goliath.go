package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zombie Goliath — Creature — Zombie Giant {4}{B}, 4/3 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eadd88b6-e75a-4482-8382-561718121772",
		Name:         "Zombie Goliath",
		Completeness: effects.CompletenessFull,
	})
}
