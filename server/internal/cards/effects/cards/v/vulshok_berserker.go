package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Berserker — Creature — Human Berserker {3}{R}, 3/2 (Mirrodin, #5606):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)"
//
// Haste is a printed keyword (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "694f7e51-7b8b-4f00-bd77-a52eded4aaa1",
		Name:            "Vulshok Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
