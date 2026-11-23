package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alabaster Host Sanctifier — Creature — Phyrexian Cleric 2/2 {1}{W} (March of the Machine, S82):
//
//	"Lifelink"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad172d4e-febf-4844-9544-861d03e27ab5",
		Name:            "Alabaster Host Sanctifier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
