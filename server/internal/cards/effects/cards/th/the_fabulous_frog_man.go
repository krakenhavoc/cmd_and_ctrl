package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Fabulous Frog-Man — Legendary Creature — Human Hero {2}{G}, 3/3
// (Marvel Super Heroes Commander, #4455): "Reach". A keyword-only body,
// the same shape as Glamorous Grapplers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0c2938d-8370-4786-8ddd-126c61538bdd",
		Name:            "The Fabulous Frog-Man",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
