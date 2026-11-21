package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Spearhunter — Creature — Lizard Warrior {2}{R}, 2/1 (Magic 2010, #5585):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// First strike is an engine keyword carried as printed data (CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf302113-7f2e-4ceb-a26f-1fd6e010be8a",
		Name:            "Viashino Spearhunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
