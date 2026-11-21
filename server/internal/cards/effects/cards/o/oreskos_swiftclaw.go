package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oreskos Swiftclaw — Creature — Cat Warrior {1}{W}, 3/1 (Journey into Nyx, #5558).
//
// A vanilla creature: the catalog entry records that nothing is missing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "014d00fc-434a-4b06-84b2-afd930677d61",
		Name:         "Oreskos Swiftclaw",
		Completeness: effects.CompletenessFull,
	})
}
