package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yoked Ox — Creature — Ox {W}, 0/4 (Theros, #5561):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19b4cf0b-e1e4-4e50-80bc-673bc6ef8bb8",
		Name:         "Yoked Ox",
		Completeness: effects.CompletenessFull,
	})
}
