package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brightblade Stoat — Creature — Weasel Soldier {1}{W}, 2/2
// (Bloomburrow, #4467):
//
//	"First strike, lifelink"
//
// Healer's Hawk's shape: both keywords ride PrintedKeywords and the
// combat engine does the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1d2f437-d474-4617-aea6-102daa5e5a15",
		Name:            "Brightblade Stoat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
	})
}
