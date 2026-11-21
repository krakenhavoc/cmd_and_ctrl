package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbtooth Wurm — Creature — Wurm {5}{G}, 6/4 (Portal Second Age,
// #5625):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03ad624e-71f6-4f72-8ce0-7700e83458d9",
		Name:         "Barbtooth Wurm",
		Completeness: effects.CompletenessFull,
	})
}
