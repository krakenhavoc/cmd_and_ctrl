package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sabretooth Tiger — Creature — Cat {2}{R}, 2/1 (Eighth Edition, #5607):
//
//	"First strike"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ea35158-1c1a-46da-b466-0eafb376e464",
		Name:            "Sabretooth Tiger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
