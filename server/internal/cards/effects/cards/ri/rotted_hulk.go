package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotted Hulk — Creature — Elemental {3}{B}, 2/5 (Journey into Nyx, #5558).
//
// A vanilla creature: the catalog entry records that nothing is missing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44fd00d5-84e4-48ec-bd6a-d53afbc2dd62",
		Name:         "Rotted Hulk",
		Completeness: effects.CompletenessFull,
	})
}
