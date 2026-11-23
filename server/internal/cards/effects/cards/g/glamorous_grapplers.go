package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glamorous Grapplers — Creature — Human Performer Villain {3}{B}, 3/3
// (Marvel Super Heroes Commander, #4455): "Menace". A vanilla menace
// body, the same shape as Apex Witchstalker's keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9f2e03fb-45c0-442d-a716-a472373c72d0",
		Name:            "Glamorous Grapplers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
