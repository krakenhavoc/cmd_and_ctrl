package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suntail Hawk — Creature — Bird {W}, 1/1 (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: only a creature with
// flying or reach blocks it (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8d31e2f-7b2e-4135-8074-9e6ef778bd80",
		Name:            "Suntail Hawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
