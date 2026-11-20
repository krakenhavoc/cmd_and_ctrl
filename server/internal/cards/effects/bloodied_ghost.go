package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodied Ghost — Creature — Spirit {1}{W/B}{W/B}, 3/3 (Eventide, #5590):
//
//	"Flying
//	 This creature enters with a -1/-1 counter on it."
//
// Flitterwing Nuisance's entry replacement (CR 614.1c): a 2/2 flyer once it
// has landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1195de01-f0ee-42c1-8e9e-a4bb90dd4932",
		Name:            "Bloodied Ghost",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 1, "Bloodied Ghost: enters with a -1/-1 counter"),
		},
	})
}
