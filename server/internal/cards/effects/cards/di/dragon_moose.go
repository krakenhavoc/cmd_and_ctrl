package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Moose — Creature — Dragon Elk {3}{R}, 3/3 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a678c6f-5a93-4a3b-80fd-4c0d0f9040c9",
		Name:            "Dragon Moose",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
