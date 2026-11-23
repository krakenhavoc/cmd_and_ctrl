package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hearthfire Hobgoblin — Creature — Goblin Soldier {R/W}{R/W}{R/W}, 2/2 (Eventide, #5590):
//
//	"Double strike"
//
// The printed keyword (CR 702.4): it deals both first-strike and regular
// combat damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc940eda-efa8-4227-9b56-0c6fbf1398d2",
		Name:            "Hearthfire Hobgoblin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
