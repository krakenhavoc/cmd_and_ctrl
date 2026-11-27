package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudcrown Oak — Creature — Treefolk Warrior {2}{G}{G}, 3/4 (Lorwyn):
//
//	"Reach"
//
// Reach is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63c52071-e3a7-4236-9dcb-e5ea68240dc2",
		Name:            "Cloudcrown Oak",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
