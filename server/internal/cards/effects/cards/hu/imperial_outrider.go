package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Outrider — Creature — Human Knight {3}{W}, 1/5 (Core Set 2020, #5518):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76cd10a0-36af-4f56-87c1-f72dee07b27c",
		Name:         "Imperial Outrider",
		Completeness: effects.CompletenessFull,
	})
}
