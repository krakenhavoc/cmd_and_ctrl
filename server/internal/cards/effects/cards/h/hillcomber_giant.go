package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hillcomber Giant — Creature — Giant Scout {2}{W}{W}, 3/3 (Lorwyn):
//
//	"Mountainwalk"
//
// Mountainwalk is the engine's: unblockable while the defending player controls
// a Mountain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a56b3434-7478-4704-b04b-92f32a1bc462",
		Name:            "Hillcomber Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
