package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hired Blade — Creature — Human Assassin {2}{B}, 3/2 (Core Set 2019, #5527):
//
//	"Flash (You may cast this spell any time you could cast an instant.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fed32c9c-7458-4143-b0b2-45ca23ba25ee",
		Name:            "Hired Blade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
