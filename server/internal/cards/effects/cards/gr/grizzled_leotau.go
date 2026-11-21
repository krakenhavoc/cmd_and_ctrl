package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzled Leotau — Creature — Cat {G}{W}, 1/5 (Alara Reborn, #5586).
//
// A vanilla creature: no rules text, so the entry only marks it reviewed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59edd0e7-37c2-45f9-9f9d-8002be459662",
		Name:         "Grizzled Leotau",
		Completeness: effects.CompletenessFull,
	})
}
