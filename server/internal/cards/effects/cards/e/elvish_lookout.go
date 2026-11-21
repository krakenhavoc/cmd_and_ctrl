package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Lookout — Creature — Elf {G}, 1/1 (Urza's Destiny, #5621):
//
//	"Shroud"
//
// One printed keyword, which the targeting check reads: no spell or
// ability can target it, its controller's included (CR 702.18a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e9aea99-1eee-4bb6-9043-a1ac2884d071",
		Name:            "Elvish Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
